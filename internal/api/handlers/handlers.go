package handlers

import (
	"allie/internal/services/arena"
	"allie/internal/services/mage"
	"allie/pkg/logger"
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

	arenaMages := h.arenaService.GetAllMages()
	arenaMagesNames := make([]string, 0, len(arenaMages))
	for _, v := range arenaMages {
		arenaMagesNames = append(arenaMagesNames, v.Username)
	}
	conn.WriteJSON(map[string]interface{}{"type": "mages", "mages": arenaMagesNames})

	var mageJoinAttempt mage.Mage
	var activeJoinedMage *mage.Mage

	for {
		var msg map[string]string
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("Read error:", err)
			break
		}

		switch msg["type"] {
		case "register":
			mageJoinAttempt.Username = msg["username"]
			mageJoinAttempt.Password, err = HashPassword(msg["password"])
			if err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
				return
			}

			err := h.mageService.Create(mageJoinAttempt)
			if err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
				continue
			}

			conn.WriteJSON(map[string]string{"type": "response", "message": "ok"})
		case "join":
			mageJoinAttempt.Username = msg["username"]
			mageJoinAttempt.Password = msg["password"]

			m, err := h.mageService.GetByUsername(mageJoinAttempt.Username)
			if err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
				continue
			}

			if CheckPasswordHash(mageJoinAttempt.Password, m.Password) {
				mageJoinAttempt.ID = m.ID
				mageJoinAttempt.HP = m.HP
				mageJoinAttempt.Conn = conn

				if activeJoinedMage != nil {
					h.arenaService.RemoveMage(activeJoinedMage.Username)
					arenaMages := h.arenaService.GetMagesExcept(activeJoinedMage.Username)
					for _, v := range arenaMages {
						v.Conn.WriteJSON(map[string]interface{}{"type": "left", "username": activeJoinedMage.Username})
					}
				}

				h.arenaService.AddMage(mageJoinAttempt)

				// mage joined the arena
				x := mageJoinAttempt
				activeJoinedMage = &x

				arenaMages := h.arenaService.GetMagesExcept(activeJoinedMage.Username)
				arenaMagesNames := make([]string, 0, len(arenaMages))
				for _, v := range arenaMages {
					v.Conn.WriteJSON(map[string]interface{}{"type": "joined", "username": activeJoinedMage.Username})
					arenaMagesNames = append(arenaMagesNames, v.Username)
				}

				conn.WriteJSON(map[string]interface{}{"type": "health", "hp": activeJoinedMage.HP, "mages": arenaMagesNames})
			} else {
				conn.WriteJSON(map[string]string{"type": "error", "message": "Invalid username/password"})
				continue
			}
		case "fireball":
			if activeJoinedMage == nil {
				conn.WriteJSON(map[string]string{"type": "error", "message": "not joined to arena"})
				continue
			}

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
			targetMage.Conn.WriteJSON(map[string]interface{}{"type": "attack", "from": activeJoinedMage.Username, "currentHP": targetMage.HP})

			if targetMage.HP <= 0 {
				targetMage.Conn.WriteJSON(map[string]interface{}{"type": "died", "by": activeJoinedMage.Username, "message": "busted"})

				h.arenaService.RemoveMage(targetMage.Username)
				arenaMages := h.arenaService.GetMagesExcept(targetMage.Username)
				for _, v := range arenaMages {
					v.Conn.WriteJSON(map[string]interface{}{"type": "died", "username": targetMage.Username, "killer": activeJoinedMage.Username})
				}

				targetMage.Conn.Close()
			}
		}
	}

	if activeJoinedMage == nil {
		return
	}

	h.arenaService.RemoveMage(activeJoinedMage.Username)
	arenaMages = h.arenaService.GetMagesExcept(activeJoinedMage.Username)
	for _, v := range arenaMages {
		v.Conn.WriteJSON(map[string]interface{}{"type": "left", "username": activeJoinedMage.Username})
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
