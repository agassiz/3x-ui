package web

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/agassiz/3x-ui/v3/internal/database"
	"github.com/agassiz/3x-ui/v3/internal/database/dbtest"
	"github.com/agassiz/3x-ui/v3/internal/database/model"
	"github.com/agassiz/3x-ui/v3/internal/sub"
	"github.com/agassiz/3x-ui/v3/internal/web/global"
	"github.com/agassiz/3x-ui/v3/internal/web/service"
)

// An email must never stand in for the subscriber's secret subId, including
// installations with a custom panel base path.
func TestEmailClashSubscriptionRouteIsUnavailable(t *testing.T) {
	for _, basePath := range []string{"/", "/private-panel/"} {
		t.Run(basePath, func(t *testing.T) {
			dbtest.InitDB(t, filepath.Join(t.TempDir(), "x-ui.db"))
			if err := (&service.SettingService{}).SetBasePath(basePath); err != nil {
				t.Fatal(err)
			}
			service.RegisterSubLinkProvider(sub.NewLinkProvider())
			db := database.GetDB()
			const uuid = "4a1b8c2d-0000-4000-8000-000000000002"
			ib := &model.Inbound{
				UserId: 1, Tag: "clash-route", Enable: true, Listen: "203.0.113.5", Port: 443,
				Protocol: model.VLESS, Remark: "clash",
				Settings:       `{"clients":[{"id":"` + uuid + `","email":"alice@e","subId":"sub-alice","enable":true}],"decryption":"none"}`,
				StreamSettings: `{"network":"tcp","security":"none","tcpSettings":{"header":{"type":"none"}}}`,
			}
			if err := db.Create(ib).Error; err != nil {
				t.Fatalf("seed inbound: %v", err)
			}
			client := &model.ClientRecord{Email: "alice@e", SubID: "sub-alice", UUID: uuid, Enable: true}
			if err := db.Create(client).Error; err != nil {
				t.Fatalf("seed client: %v", err)
			}
			if err := db.Create(&model.ClientInbound{ClientId: client.Id, InboundId: ib.Id}).Error; err != nil {
				t.Fatalf("seed client_inbound: %v", err)
			}

			previous := global.GetWebServer()
			s := NewServer()
			s.cron = cron.New(cron.WithLocation(time.Local), cron.WithSeconds())
			global.SetWebServer(s)
			t.Cleanup(func() {
				s.cancel()
				global.SetWebServer(previous)
			})
			engine, err := s.initRouter()
			if err != nil {
				t.Fatalf("init router: %v", err)
			}

			for _, email := range []string{"alice@e", "nobody@e"} {
				for _, method := range []string{http.MethodGet, http.MethodHead} {
					rec := httptest.NewRecorder()
					req := httptest.NewRequest(method, basePath+"clash/subscription/"+email, nil)
					engine.ServeHTTP(rec, req)
					if rec.Code != http.StatusNotFound {
						t.Fatalf("%s %s: status = %d, want 404", method, req.URL.Path, rec.Code)
					}
					if strings.Contains(rec.Body.String(), uuid) {
						t.Fatal("email route exposed the proxy credential")
					}
				}
			}
		})
	}
}
