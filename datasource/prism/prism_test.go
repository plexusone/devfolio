package prism

import (
	"strings"
	"testing"
)

const sampleJSONL = `{"kind":"initiative","initiative":{"id":"INIT-PRISMCONTROL-001","title":"PRISM Control MVP","description":"Build the canonical product delivery control plane","status":"in_progress"},"exportedAt":"2026-07-20T10:00:00Z"}
{"kind":"rmi","rmi":{"id":"RMI-PRISMCONTROL-001","repo":"github.com/ProductBuildersHQ/prism-control","initiative":"INIT-PRISMCONTROL-001","phase":"INIT-PRISMCONTROL-001/phase-1","title":"Ent schema for initiatives and RMIs","type":"capability","status":"completed","required":true,"completedAt":"2026-06-15T12:00:00Z"},"exportedAt":"2026-07-20T10:00:00Z"}
{"kind":"rmi","rmi":{"id":"RMI-PRISMCONTROL-005","repo":"github.com/ProductBuildersHQ/prism-control","initiative":"INIT-PRISMCONTROL-001","phase":"INIT-PRISMCONTROL-001/phase-1","title":"Unit-of-work with Dolt commit","type":"capability","status":"in_progress","required":true,"assignedTo":"session-1234"},"exportedAt":"2026-07-20T10:00:00Z"}
{"kind":"initiative","initiative":{"id":"INIT-DEVFOLIO-001","title":"DevFolio MVP","status":"planned"},"exportedAt":"2026-07-20T10:00:00Z"}
{"kind":"rmi","rmi":{"id":"RMI-DEVFOLIO-001","repo":"github.com/plexusone/devfolio","initiative":"INIT-DEVFOLIO-001","phase":"INIT-DEVFOLIO-001/phase-1","title":"Initiative dimension datasource","type":"capability","status":"proposed","required":false},"exportedAt":"2026-07-20T10:00:00Z"}`

func TestLoad(t *testing.T) {
	data, err := Load(strings.NewReader(sampleJSONL))
	if err != nil {
		t.Fatal(err)
	}

	if len(data.Initiatives) != 2 {
		t.Fatalf("expected 2 initiatives, got %d", len(data.Initiatives))
	}
	if len(data.RMIs) != 3 {
		t.Fatalf("expected 3 RMIs, got %d", len(data.RMIs))
	}

	// Verify first initiative
	init1 := data.Initiatives[0]
	if init1.ID != "INIT-PRISMCONTROL-001" {
		t.Errorf("initiative[0].ID: got %q, want %q", init1.ID, "INIT-PRISMCONTROL-001")
	}
	if init1.Title != "PRISM Control MVP" {
		t.Errorf("initiative[0].Title: got %q, want %q", init1.Title, "PRISM Control MVP")
	}
	if init1.Status != "in_progress" {
		t.Errorf("initiative[0].Status: got %q, want %q", init1.Status, "in_progress")
	}

	// Verify RMI with assignment
	rmi2 := data.RMIs[1]
	if rmi2.ID != "RMI-PRISMCONTROL-005" {
		t.Errorf("rmi[1].ID: got %q, want %q", rmi2.ID, "RMI-PRISMCONTROL-005")
	}
	if rmi2.AssignedTo != "session-1234" {
		t.Errorf("rmi[1].AssignedTo: got %q, want %q", rmi2.AssignedTo, "session-1234")
	}
	if rmi2.Status != "in_progress" {
		t.Errorf("rmi[1].Status: got %q, want %q", rmi2.Status, "in_progress")
	}

	// Verify exportedAt was parsed
	if data.ExportedAt.IsZero() {
		t.Error("expected ExportedAt to be set")
	}
	if data.ExportedAt.Year() != 2026 {
		t.Errorf("ExportedAt year: got %d, want 2026", data.ExportedAt.Year())
	}
}

func TestRMIsByInitiative(t *testing.T) {
	data, err := Load(strings.NewReader(sampleJSONL))
	if err != nil {
		t.Fatal(err)
	}

	byInit := data.RMIsByInitiative()
	prismRMIs := byInit["INIT-PRISMCONTROL-001"]
	if len(prismRMIs) != 2 {
		t.Fatalf("expected 2 RMIs for INIT-PRISMCONTROL-001, got %d", len(prismRMIs))
	}

	devfolioRMIs := byInit["INIT-DEVFOLIO-001"]
	if len(devfolioRMIs) != 1 {
		t.Fatalf("expected 1 RMI for INIT-DEVFOLIO-001, got %d", len(devfolioRMIs))
	}
}

func TestRMIsByRepo(t *testing.T) {
	data, err := Load(strings.NewReader(sampleJSONL))
	if err != nil {
		t.Fatal(err)
	}

	byRepo := data.RMIsByRepo()
	prismRMIs := byRepo["github.com/ProductBuildersHQ/prism-control"]
	if len(prismRMIs) != 2 {
		t.Fatalf("expected 2 RMIs for prism-control, got %d", len(prismRMIs))
	}

	devfolioRMIs := byRepo["github.com/plexusone/devfolio"]
	if len(devfolioRMIs) != 1 {
		t.Fatalf("expected 1 RMI for devfolio, got %d", len(devfolioRMIs))
	}
}

func TestLoadEmptyLines(t *testing.T) {
	input := `{"kind":"initiative","initiative":{"id":"INIT-X-001","title":"Test","status":"planned"}}

{"kind":"rmi","rmi":{"id":"RMI-X-001","repo":"github.com/org/repo","initiative":"INIT-X-001","phase":"INIT-X-001/phase-1","title":"Task","type":"task","status":"proposed","required":true}}`

	data, err := Load(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(data.Initiatives) != 1 {
		t.Errorf("expected 1 initiative, got %d", len(data.Initiatives))
	}
	if len(data.RMIs) != 1 {
		t.Errorf("expected 1 RMI, got %d", len(data.RMIs))
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	_, err := Load(strings.NewReader(`not valid json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoadUnknownKind(t *testing.T) {
	_, err := Load(strings.NewReader(`{"kind":"unknown"}`))
	if err == nil {
		t.Fatal("expected error for unknown kind, got nil")
	}
}

func TestLoadNilInitiative(t *testing.T) {
	_, err := Load(strings.NewReader(`{"kind":"initiative"}`))
	if err == nil {
		t.Fatal("expected error for nil initiative field, got nil")
	}
}

func TestLoadNilRMI(t *testing.T) {
	_, err := Load(strings.NewReader(`{"kind":"rmi"}`))
	if err == nil {
		t.Fatal("expected error for nil rmi field, got nil")
	}
}
