// context_hijacker.go — context-hijacker.mjs Go 포팅
// MITM HTTPS 프록시: 타겟 호스트의 API 트래픽을 캡처
// 외부 의존성: 0 (Go stdlib only)
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var mitmTargetHosts = []string{
	"cloudcode-pa.googleapis.com",
	"generativelanguage.googleapis.com",
	"anthropic.com",
}

var mitmCaptureCount int

func isMitmTarget(host string) bool {
	for _, t := range mitmTargetHosts {
		if strings.Contains(host, t) {
			return true
		}
	}
	return false
}

func mitmCapture(body []byte, url, method, dumpDir, outputFile string) {
	if len(body) <= 10 {
		return
	}
	mitmCaptureCount++
	dumpFile := filepath.Join(dumpDir, fmt.Sprintf("raw_dump_%d_%d.bin", time.Now().UnixMilli(), mitmCaptureCount))
	os.WriteFile(dumpFile, body, 0600)

	md := fmt.Sprintf("# Hijacked Context — %s\n\nURL: %s\n\n```text\n%s\n```\n",
		time.Now().Format(time.RFC3339), url, truncStr(string(body), 50000))
	os.WriteFile(outputFile, []byte(md), 0600)
	fmt.Printf("[MITM CAPTURE] Dumped %d bytes from %s %s\n", len(body), method, url)
}

func truncStr(s string, max int) string {
	if len(s) > max {
		return s[:max]
	}
	return s
}

// generateSelfSignedCert generates an in-memory TLS cert for MITM
func generateMitmCert() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{Organization: []string{"NeuronFS MITM"}},
		DNSNames:     append(mitmTargetHosts, "*.googleapis.com", "gemini.googleapis.com"),
		NotBefore:    time.Now().Add(-1 * time.Hour),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, _ := x509.MarshalECPrivateKey(key)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	return tls.X509KeyPair(certPEM, keyPEM)
}

func runContextHijacker(brainRoot string) {
	inboxDir := filepath.Join(brainRoot, "_agents", "global_inbox")
	dumpDir := filepath.Join(inboxDir, "grpc_dumps")
	outputFile := filepath.Join(inboxDir, "latest_hijacked_context.md")
	os.MkdirAll(dumpDir, 0750)

	cert, err := generateMitmCert()
	if err != nil {
		fmt.Printf("[MITM] cert error: %v\n", err)
		return
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"http/1.1"}, // Force HTTP/1.1
	}

	// MITM TLS termination server (random port)
	mitmListener, err := tls.Listen("tcp", "127.0.0.1:0", tlsConfig)
	if err != nil {
		fmt.Printf("[MITM] listen error: %v\n", err)
		return
	}
	mitmPort := mitmListener.Addr().(*net.TCPAddr).Port
	fmt.Printf("[MITM] TLS termination on :%d\n", mitmPort)

	// Handle MITM connections
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			urlStr := "https://" + r.Host + r.URL.String()

			if isMitmTarget(r.Host) {
				mitmCapture(body, urlStr, r.Method, dumpDir, outputFile)
			}

			// Forward upstream
			tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
			upReq, _ := http.NewRequest(r.Method, urlStr, strings.NewReader(string(body)))
			for k, vv := range r.Header {
				for _, v := range vv {
					upReq.Header.Add(k, v)
				}
			}
			resp, err := tr.RoundTrip(upReq)
			if err != nil {
				w.WriteHeader(502)
				return
			}
			defer resp.Body.Close()
			for k, vv := range resp.Header {
				for _, v := range vv {
					w.Header().Add(k, v)
				}
			}
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
		})
		http.Serve(mitmListener, mux)
	}()

	// Raw TCP proxy on :8080 — handles CONNECT + plain HTTP
	listener, err := net.Listen("tcp", "127.0.0.1:8080")
	if err != nil {
		fmt.Printf("[MITM] proxy listen error: %v\n", err)
		return
	}

	fmt.Println("📡 NeuronFS Context Hijacker (Go Native)")
	fmt.Println("   Listening: 127.0.0.1:8080")
	fmt.Printf("   Targets: %s\n", strings.Join(mitmTargetHosts, " | "))

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go handleMitmConn(conn, mitmPort)
	}
}

func handleMitmConn(clientConn net.Conn, mitmPort int) {
	defer clientConn.Close()

	buf := make([]byte, 4096)
	n, err := clientConn.Read(buf)
	if err != nil {
		return
	}

	req := string(buf[:n])
	if !strings.HasPrefix(req, "CONNECT ") {
		// Plain HTTP — not handled, just close
		return
	}

	hostPort := strings.Split(strings.Fields(req)[1], ":")
	host := hostPort[0]

	// ACK the CONNECT
	clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	var targetConn net.Conn
	if isMitmTarget(host) {
		targetConn, err = net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", mitmPort), 3*time.Second)
	} else {
		port := "443"
		if len(hostPort) > 1 {
			port = hostPort[1]
		}
		targetConn, err = net.DialTimeout("tcp", host+":"+port, 5*time.Second)
	}
	if err != nil {
		return
	}
	defer targetConn.Close()

	go io.Copy(targetConn, clientConn)
	io.Copy(clientConn, targetConn)
}
