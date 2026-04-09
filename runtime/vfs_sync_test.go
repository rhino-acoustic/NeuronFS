package main

import (
	"testing"
	"time"
)

func TestSyncToNAS_StopChannel(t *testing.T) {
	// SyncToNAS가 stopCh로 즉시 종료되는지 검증
	stopCh := make(chan struct{})
	close(stopCh) // 즉시 종료 신호

	done := make(chan bool, 1)
	go func() {
		SyncToNAS(t.TempDir(), t.TempDir(), stopCh)
		done <- true
	}()

	select {
	case <-done:
		// 정상 종료
	case <-time.After(3 * time.Second):
		t.Fatal("SyncToNAS did not stop within 3s after stopCh closed")
	}
}

func TestSyncToNAS_InvalidPaths(t *testing.T) {
	// 존재하지 않는 경로에서도 패닉 없이 실행
	stopCh := make(chan struct{})

	done := make(chan bool, 1)
	go func() {
		// 한 번 robocopy 실행 후 즉시 종료
		go func() {
			time.Sleep(500 * time.Millisecond)
			close(stopCh)
		}()
		SyncToNAS("/nonexistent/src", "/nonexistent/dst", stopCh)
		done <- true
	}()

	select {
	case <-done:
		// 정상
	case <-time.After(10 * time.Second):
		t.Fatal("SyncToNAS hung on invalid paths")
	}
}
