package handlers

import (
	"allie/pkg/logger"
	"allie/pkg/services/arena"
	"allie/pkg/services/mage"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"go.uber.org/fx"
	"golang.org/x/crypto/bcrypt"
)

var Module = fx.Provide(New)

type Handlers interface {
	HandleWebsocket(w http.ResponseWriter, r *http.Request)
}

type handlers struct {
	logger       logger.Logger
	mageService  mage.Service
	arenaService arena.Service
}

type Params struct {
	fx.In
	Logger       logger.Logger
	MageService  mage.Service
	ArenaService arena.Service
}

func New(p Params) Handlers {
	return &handlers{
		logger:       p.Logger,
		mageService:  p.MageService,
		arenaService: p.ArenaService,
	}
}

var upgrader = websocket.Upgrader{}

func (h *handlers) HandleWebsocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	mages := h.arenaService.GetAllMages()
	magesResp := make([]string, 0, len(mages))
	for _, v := range mages {
		magesResp = append(magesResp, v.Username)
	}
	conn.WriteJSON(map[string]interface{}{"type": "mages", "mages": magesResp})

	var mageInfo mage.Mage
	var lastArenaJoin *mage.Mage

	for {
		var msg map[string]string
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("Read error:", err)
			break
		}

		switch msg["type"] {
		case "register":
			mageInfo.Username = msg["username"]
			mageInfo.Password = msg["password"]

			err := h.mageService.Create(mageInfo)
			if err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
				continue
			}

			conn.WriteJSON(map[string]string{"type": "response", "message": "ok"})
		case "join":
			mageInfo.Username = msg["username"]
			mageInfo.Password = msg["password"]

			m, err := h.mageService.GetByUsername(mageInfo.Username)
			if err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
				continue
			}

			if mageInfo.Password == m.Password {
				mageInfo.ID = m.ID
				mageInfo.HP = m.HP
				mageInfo.Conn = conn

				h.arenaService.AddMage(mageInfo)

				if lastArenaJoin != nil {
					h.arenaService.RemoveMage(lastArenaJoin.Username)
					mages := h.arenaService.GetMagesFor(lastArenaJoin.Username)
					for _, v := range mages {
						v.Conn.WriteJSON(map[string]interface{}{"type": "left", "username": mageInfo.Username})
					}
				}

				lastArenaJoin = &mageInfo

				mages := h.arenaService.GetMagesFor(mageInfo.Username)
				magesResp := make([]string, 0, len(mages))
				for _, v := range mages {
					v.Conn.WriteJSON(map[string]interface{}{"type": "joined", "username": mageInfo.Username})
					magesResp = append(magesResp, v.Username)
				}

				conn.WriteJSON(map[string]interface{}{"type": "health", "hp": mageInfo.HP, "mages": magesResp})
			} else {
				conn.WriteJSON(map[string]string{"type": "error", "message": "Invalid username/password"})
				return
			}
		case "fireball":
			from := msg["from"]
			target := msg["target"]

			targetMage := h.arenaService.GetByUsername(target)
			if targetMage == nil {
				conn.WriteJSON(map[string]string{"type": "error", "message": "no target found"})
				continue
			}

			err = h.mageService.ReduceHPByUsername(target)
			if err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
				continue
			}

			targetMage.HP -= 10
			targetMage.Conn.WriteJSON(map[string]interface{}{"type": "attack", "from": from, "currentHP": targetMage.HP})

			if targetMage.HP <= 0 {
				targetMage.Conn.WriteJSON(map[string]interface{}{"type": "died", "by": from, "message": "busted"})

				h.arenaService.RemoveMage(targetMage.Username)
				mages := h.arenaService.GetMagesFor(targetMage.Username)
				for _, v := range mages {
					v.Conn.WriteJSON(map[string]interface{}{"type": "died", "username": targetMage.Username})
				}

				targetMage.Conn.Close()
			}
		}
	}

	if lastArenaJoin == nil {
		return
	}

	h.arenaService.RemoveMage(mageInfo.Username)
	mages = h.arenaService.GetMagesFor(mageInfo.Username)
	for _, v := range mages {
		v.Conn.WriteJSON(map[string]interface{}{"type": "left", "username": mageInfo.Username})
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
