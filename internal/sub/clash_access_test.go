package sub

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/agassiz/3x-ui/v3/internal/database"
	"github.com/agassiz/3x-ui/v3/internal/database/model"
)

func TestClashSubscriptionRequiresCurrentSubID(t *testing.T) {
	router, subID := initHwidSubRouter(t, 0)
	const uuid = "11111111-2222-4333-8444-555555555555"
	rec := requestSub(t, router, http.MethodGet, "/clash/"+subID, "", "application/yaml")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), uuid) {
		t.Fatalf("valid subscription: status=%d body=%s", rec.Code, rec.Body.String())
	}
	if err := database.GetDB().Model(&model.ClientRecord{}).Where("sub_id = ?", subID).
		Update("sub_id", "rotated-token").Error; err != nil {
		t.Fatal(err)
	}
	for _, oldID := range []string{subID, "route@example.com"} {
		rec = requestSub(t, router, http.MethodGet, "/clash/"+oldID, "", "application/yaml")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("old token or email %q: status=%d, want 404", oldID, rec.Code)
		}
	}
	rec = requestSub(t, router, http.MethodGet, "/clash/rotated-token", "", "application/yaml")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), uuid) {
		t.Fatalf("rotated subscription: status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestClashSubscriptionDisabledDoesNotExposeFormatRoutes(t *testing.T) {
	_, subID := initHwidSubRouter(t, 0)
	router := gin.New()
	NewSUBController(router.Group("/"), WithSUBClashEnabled(false))
	for _, path := range []string{"/clash/", "/mihomo/", "/clash-legacy/"} {
		rec := requestSub(t, router, http.MethodGet, path+subID, "", "application/yaml")
		if rec.Code != http.StatusNotFound {
			t.Fatalf("disabled route %q: status=%d, want 404", path, rec.Code)
		}
	}
}

func TestSubscriptionDisabledDoesNotStartListener(t *testing.T) {
	initHwidSubRouter(t, 0)
	db := database.GetDB()
	if err := db.Where("key = ?", "subEnable").Delete(&model.Setting{}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Setting{Key: "subEnable", Value: "false"}).Error; err != nil {
		t.Fatal(err)
	}
	s := NewServer()
	t.Cleanup(s.cancel)
	if err := s.Start(); err != nil {
		t.Fatal(err)
	}
	if s.listener != nil || s.httpServer != nil {
		t.Fatal("disabled subscription service opened a listener")
	}
}
