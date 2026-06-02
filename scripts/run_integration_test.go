package scripts_test

import (
	"os"
	"strings"
	"testing"
)

func TestRunIntegrationGeneratesStandardImpactBeforeReleaseEvidence(t *testing.T) {
	contents, err := os.ReadFile("run_integration.sh")
	if err != nil {
		t.Fatalf("read run_integration.sh: %v", err)
	}

	script := string(contents)
	standardImpactIndex := strings.Index(script, "GOWORK=off make standard-impact-check")
	if standardImpactIndex < 0 {
		t.Fatal("run_integration.sh does not generate standard impact evidence")
	}
	evidenceIndex := strings.Index(script, "CHECK_STATUS=passed GOWORK=off make evidence")
	if evidenceIndex < 0 {
		t.Fatal("run_integration.sh does not generate release evidence")
	}
	checkIndex := strings.Index(script, "RELEASE_EVIDENCE_REQUIRE_PASSED=1 GOWORK=off make release-evidence-check")
	if checkIndex < 0 {
		t.Fatal("run_integration.sh does not verify release evidence")
	}

	if standardImpactIndex > evidenceIndex || evidenceIndex > checkIndex {
		t.Fatalf("integration evidence order is wrong: standard-impact=%d evidence=%d release-check=%d", standardImpactIndex, evidenceIndex, checkIndex)
	}
}
