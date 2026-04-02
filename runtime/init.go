package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// ?€?€?€ Brain Initializer ?€?€?€
// Creates the correct folder-as-neuron structure
// Each call to neuron() creates: folder + N.neuron counter file

type brainInit struct {
	root string
}

// neuron simplifies creating a Neuron struct for default system initialization.
func (b *brainInit) neuron(path string, counter int, signals ...string) {
	dir := filepath.Join(b.root, path)
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, fmt.Sprintf("%d.neuron", counter)), []byte(""), 0644)
	for _, sig := range signals {
		os.WriteFile(filepath.Join(dir, sig+".neuron"), []byte(""), 0644)
	}
}

// axon converts a source/target rule into an Axon struct for cross-references.
func (b *brainInit) axon(path, target string) {
	full := filepath.Join(b.root, path)
	os.WriteFile(full, []byte("TARGET: "+target), 0644)
}

// initBrain initializes the brain directory structure and injects default rules if missing.
func initBrain(root string) {
	if _, err := os.Stat(root); err == nil {
		fmt.Printf("[CLEAN] Removing %s\n", root)
		os.RemoveAll(root)
	}

	b := &brainInit{root: root}
	fmt.Println("=== NeuronFS Brain v4 Init ===")

	// ?â”??BRAINSTEM ?â”??	fmt.Println("[1/7] brainstem")
	b.neuron("brainstem/no_simulation_real_results", 99)
	b.neuron("brainstem/no_repeat_same_mistakes", 99)
	b.neuron("brainstem/never_use_fallback", 99)
	b.neuron("brainstem/quality_over_speed", 99)
	b.neuron("brainstem/self_debug_visual_verify", 99)
	b.neuron("brainstem/execute_dont_discuss", 99)
	b.neuron("brainstem/verify_before_deliver", 99)           // ê²€ì¦????„ë‹¬
	b.neuron("brainstem/auto_iterate_until_satisfied", 99)    // ?ê¸°_ë°˜ë³µ_ê²€ì¦?	b.neuron("brainstem/no_hardcoding", 99)                   // ?˜ë“œì½”ë”©_ê¸ˆì?
	b.neuron("brainstem/no_process_bypass", 99)               // ?„ë¡œ?¸ìŠ¤_?°íšŒ_ê¸ˆì?
	b.neuron("brainstem/understand_direction_accurately", 99) // ë°©í–¥_?•í™•_?´í•´
	b.neuron("brainstem/always_use_async_await", 50)
	b.neuron("brainstem/no_ampersand_use_semicolon", 50)
	b.neuron("brainstem/scripts_must_be_ps1", 50)

	// ?â”??LIMBIC ?â”??	fmt.Println("[2/7] limbic")
	b.neuron("limbic/detect_frustration", 1)
	b.neuron("limbic/detect_urgency", 1)
	b.neuron("limbic/detect_praise", 1)
	b.neuron("limbic/adrenaline_emergency", 1)
	b.neuron("limbic/dopamine_reward", 1, "dopamine1")
	b.neuron("limbic/endorphin_persistence", 1)
	b.neuron("limbic/strip_emotion_forward_goal", 1)

	// ?â”??HIPPOCAMPUS ?â”??	fmt.Println("[3/7] hippocampus")
	b.neuron("hippocampus/session_log", 1, "memory1")
	b.neuron("hippocampus/bomb_history", 1, "memory1")
	b.neuron("hippocampus/error_patterns", 1, "memory1")
	b.neuron("hippocampus/dopamine_log", 1, "memory1")
	b.neuron("hippocampus/context_restore_from_previous", 30, "memory1")  // ?€??ë³µì›
	b.neuron("hippocampus/user_correction_ground_truth", 30, "memory1")   // ?¬ìš©???œë³¸_?™ìŠµ

	// ?â”??SENSORS ?â”??	fmt.Println("[4/7] sensors")
	b.neuron("sensors/nas/write_cmd_copy_only", 30)
	b.neuron("sensors/nas/no_powershell_copyitem", 30)
	b.neuron("sensors/nas/test_path_before_write", 20)
	b.neuron("sensors/nas/robocopy_for_large_files", 10)
	b.neuron("sensors/design/sandstone_base_f7f1e8", 20)
	b.neuron("sensors/design/glassmorphism_blur20", 15)
	b.neuron("sensors/design/button_rounded_full", 15)
	b.neuron("sensors/typography/font_suit_ko_instrument_en", 20)
	b.neuron("sensors/brand/vegavery_run_premium_wellness", 30)    // ë¸Œëžœ???•ì²´??	b.neuron("sensors/brand/tone_premium_natural_luxury", 30)      // ?¤ì•¤ë§¤ë„ˆ
	b.neuron("sensors/nas_brain/path_z_vol1_vgvr_brain_lw", 20)   // NAS BRAIN ê²½ë¡œ

	// ?â”??CORTEX ?â”??	fmt.Println("[5/7] cortex")

	// frontend
	b.neuron("cortex/frontend/css/glass_blur20_alpha15", 10)
	b.neuron("cortex/frontend/css/section_gap_80_128px", 10)
	b.neuron("cortex/frontend/css/accent_blue_3b82f6", 10)
	b.neuron("cortex/frontend/css/primary_sandstone", 10)
	b.neuron("cortex/frontend/css/rounded_50px_dark", 8)
	b.neuron("cortex/frontend/css/fade_in_up_06s", 8)
	b.neuron("cortex/frontend/css/stagger_100ms", 8)
	b.neuron("cortex/frontend/typography/instrument_serif_italic", 12)
	b.neuron("cortex/frontend/typography/suit_400_700", 12)
	b.neuron("cortex/frontend/react/hooks_pattern", 15)
	b.neuron("cortex/frontend/coding/comment_every_selector", 10)

	// backend
	b.neuron("cortex/backend/supabase/rls_always_on", 20)
	b.neuron("cortex/backend/devops/multi_stage_build", 5)

	// methodology (from global principles)
	b.neuron("cortex/methodology/community_academic_search", 30)   // ì»¤ë??ˆí‹°_?™ê³„_ê²€??	b.neuron("cortex/methodology/positive_negative_both", 30)      // ê¸ì •_ë¶€???‘ë°©??	b.neuron("cortex/methodology/two_persona_debate", 20)          // ???˜ë¥´?Œë‚˜_?¼ìŸ
	b.neuron("cortex/methodology/third_party_audit", 20)           // ?????œì„ _ê°ì‚¬
	b.neuron("cortex/methodology/dictionary_based_matching", 20)   // ?¬ì „_ê¸°ë°˜_ë§¤ì¹­
	b.neuron("cortex/methodology/ask_only_when_necessary", 30)     // ?„ìš”??ê²½ìš°ë§?ì§ˆë¬¸

	// NeuronFS meta-knowledge
	b.neuron("cortex/neuronfs/axiom/folder_is_neuron", 99, "dopamine1")
	b.neuron("cortex/neuronfs/axiom/file_is_firing_trace", 99, "dopamine1")
	b.neuron("cortex/neuronfs/axiom/path_is_sentence", 99, "dopamine1")
	b.neuron("cortex/neuronfs/axiom/counter_is_activation", 50)
	b.neuron("cortex/neuronfs/axiom/depth_is_specificity", 50)

	b.neuron("cortex/neuronfs/signals/bomb_circuit_breaker", 30)
	b.neuron("cortex/neuronfs/signals/dopamine_reinforcement", 30)
	b.neuron("cortex/neuronfs/signals/dormant_pruning", 10)
	b.neuron("cortex/neuronfs/signals/counter_as_filename", 50)

	b.neuron("cortex/neuronfs/structure/subsumption_priority", 50)
	b.neuron("cortex/neuronfs/structure/small_world_network", 30)
	b.neuron("cortex/neuronfs/structure/axon_crosslink", 30)
	b.neuron("cortex/neuronfs/structure/seven_regions", 50)

	b.neuron("cortex/neuronfs/growth/experience_only_division", 99)
	b.neuron("cortex/neuronfs/growth/synapse_explosion", 20)
	b.neuron("cortex/neuronfs/growth/pruning_dormant", 20)
	b.neuron("cortex/neuronfs/growth/myelination_highway", 20)
	b.neuron("cortex/neuronfs/growth/brainstem_lock_maturity", 30)
	b.neuron("cortex/neuronfs/growth/folder_hierarchy_unlimited_depth", 50)

	b.neuron("cortex/neuronfs/runtime/scanner_reads_tree", 30)
	b.neuron("cortex/neuronfs/runtime/compiler_path_to_sentence", 30)
	b.neuron("cortex/neuronfs/runtime/injector_to_gemini", 30)
	b.neuron("cortex/neuronfs/runtime/counter_writeback", 30)

	b.neuron("cortex/neuronfs/wargame/folder_equals_neuron_18of20", 99)
	b.neuron("cortex/neuronfs/wargame/file_equals_trace_16of20", 80)
	b.neuron("cortex/neuronfs/wargame/axon_crosslink_14of20", 70)
	b.neuron("cortex/neuronfs/wargame/counter_activation_13of20", 65)
	b.neuron("cortex/neuronfs/wargame/router_spotlight_12of20", 60)
	b.neuron("cortex/neuronfs/wargame/bomb_pain_11of20", 55)
	b.neuron("cortex/neuronfs/wargame/brainstem_conscience_10of20", 50)

	b.neuron("cortex/neuronfs/connections/permanent_manual", 20)
	b.neuron("cortex/neuronfs/connections/router_assigned_auto", 20)
	b.neuron("cortex/neuronfs/connections/tunneled_temporary", 20)

	b.neuron("cortex/neuronfs/ownership/brainstem_human_only", 50)
	b.neuron("cortex/neuronfs/ownership/limbic_system_auto", 50)
	b.neuron("cortex/neuronfs/ownership/cortex_ai_experience", 50)
	b.neuron("cortex/neuronfs/ownership/hippocampus_auto_log", 50)
	b.neuron("cortex/neuronfs/ownership/sensors_human_declare", 50)
	b.neuron("cortex/neuronfs/ownership/ego_human_customize", 50)
	b.neuron("cortex/neuronfs/ownership/prefrontal_human_set", 50)

	b.neuron("cortex/neuronfs/defense/brainstem_readonly_lock", 30)
	b.neuron("cortex/neuronfs/defense/server_db_snapshot", 10)
	b.neuron("cortex/neuronfs/defense/bomb_circuit_breaker_auto", 30)

	// ?â”??CORTEX/SKILLS ??External skill references ?â”??	home := os.Getenv("USERPROFILE")
	if home == "" {
		home = os.Getenv("HOME") // fallback for non-Windows
	}
	skillBase := filepath.Join(home, ".agents", "skills")

	b.neuron("cortex/skills/supanova/taste_skill", 20)
	b.axon("cortex/skills/supanova/taste_skill/ref.axon", "SKILL:"+filepath.Join(skillBase, "taste-skill", "SKILL.md"))
	b.neuron("cortex/skills/supanova/redesign_skill", 15)
	b.axon("cortex/skills/supanova/redesign_skill/ref.axon", "SKILL:"+filepath.Join(skillBase, "redesign-skill", "SKILL.md"))
	b.neuron("cortex/skills/supanova/soft_skill", 15)
	b.axon("cortex/skills/supanova/soft_skill/ref.axon", "SKILL:"+filepath.Join(skillBase, "soft-skill", "SKILL.md"))
	b.neuron("cortex/skills/supanova/output_skill", 15)
	b.axon("cortex/skills/supanova/output_skill/ref.axon", "SKILL:"+filepath.Join(skillBase, "output-skill", "SKILL.md"))

	// ?â”??EGO ?â”??	fmt.Println("[6/7] ego")
	b.neuron("ego/expert_concise", 30)
	b.neuron("ego/korean_native", 30)
	b.neuron("ego/transistor_gate_decomposition", 20)
	b.neuron("ego/opus_discover_then_delegate", 20)
	b.neuron("ego/community_verified_methods", 15)
	b.neuron("ego/aggressive_rebuild", 10)
	b.neuron("ego/conservative_patch", 10)

	// ?â”??PREFRONTAL ?â”??	fmt.Println("[7/7] prefrontal")
	b.neuron("prefrontal/long_term_direction", 1)
	b.neuron("prefrontal/current_sprint", 1)
	b.neuron("prefrontal/future_tasks", 1)
	b.neuron("prefrontal/project/omniverse_market_research", 20)
	b.neuron("prefrontal/project/neuronfs_brain_evolution", 50)
	b.neuron("prefrontal/project/vegavery_crm_operations", 15)
	b.neuron("prefrontal/project/video_pipeline_v17", 10)

	// ?â”??AXON crosslinks ??Layered Network ?â”??	// Subsumption cascade: brainstem ??limbic ??hippocampus ??sensors ??cortex ??ego ??prefrontal
	// Each layer checks the layer above before acting (priority = layer order)
	fmt.Println("[AXON] layered network")

	// --- Cascade (top-down priority chain) ---
	b.axon("brainstem/cascade_to_limbic.axon", "limbic")          // bomb?´ë©´ ê°ì • ì°¨ë‹¨
	b.axon("limbic/cascade_from_brainstem.axon", "brainstem")     // ê°ì • ?„ì— ?‘ì‹¬ ì²´í¬
	b.axon("limbic/cascade_to_hippocampus.axon", "hippocampus")   // ê°ì •??ê¸°ì–µ ?¸ë¦¬ê±?	b.axon("hippocampus/cascade_from_limbic.axon", "limbic")      // ê¸°ì–µ ?„ì— ê°ì • ì²´í¬
	b.axon("hippocampus/cascade_to_sensors.axon", "sensors")      // ê¸°ì–µ???˜ê²½ ?¸ì‹???í–¥
	b.axon("sensors/cascade_from_hippocampus.axon", "hippocampus")// ?˜ê²½ ?„ì— ê³¼ê±° ?¨í„´ ì²´í¬
	b.axon("sensors/cascade_to_cortex.axon", "cortex")            // ?˜ê²½ ?œì•½??ì§€???„í„°ë§?	b.axon("cortex/cascade_from_sensors.axon", "sensors")         // ì§€???ìš© ?„ì— ?˜ê²½ ?•ì¸
	b.axon("cortex/cascade_to_ego.axon", "ego")                   // ì§€?ì´ ?œí˜„ ë°©ì‹ ê²°ì •
	b.axon("ego/cascade_from_cortex.axon", "cortex")              // ??ê²°ì • ?„ì— ì§€???•ì¸
	b.axon("ego/cascade_to_prefrontal.axon", "prefrontal")        // ?±í–¥??ëª©í‘œ ?´ì„???í–¥
	b.axon("prefrontal/cascade_from_ego.axon", "ego")             // ëª©í‘œ ?„ì— ?±í–¥ ?•ì¸

	// --- Cross-links (shortcuts = small-world network) ---
	b.axon("prefrontal/shortcut_to_cortex.axon", "cortex")       // ëª©í‘œê°€ ì§ì ‘ ì§€??? íƒ
	b.axon("cortex/shortcut_to_hippocampus.axon", "hippocampus") // ?™ìŠµ ê²°ê³¼ë¥?ê¸°ì–µ??ê¸°ë¡
	b.axon("limbic/shortcut_to_cortex.axon", "cortex")           // ê¸´ê¸‰ ??ì§€??ì§ì ‘ ?‘ê·¼
	b.axon("sensors/shortcut_to_brainstem.axon", "brainstem")    // ?˜ê²½ ?„í—˜ ??ë³¸ëŠ¥ ì§ì ‘ ë°œë™

	// Stats
	neuronCount := 0
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}
		files, _ := filepath.Glob(filepath.Join(path, "*.neuron"))
		if len(files) > 0 {
			neuronCount++
		}
		return nil
	})

	fmt.Println("\n=== COMPLETE ===")
	fmt.Printf("  Root: %s\n", root)
	fmt.Printf("  Neurons (folders): %d\n", neuronCount)
}

