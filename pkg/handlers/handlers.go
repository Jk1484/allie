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

	arenaMages := h.arenaService.GetAllMages()
	arenaMagesNames := make([]string, 0, len(arenaMages))
	for _, v := range arenaMages {
		arenaMagesNames = append(arenaMagesNames, v.Username)
	}
	conn.WriteJSON(map[string]interface{}{"type": "mages", "mages": arenaMagesNames})

	var currentMage mage.Mage
	var lastJoinedMage *mage.Mage

	for {
		var msg map[string]string
		if err := conn.ReadJSON(&msg); err != nil {
			log.Println("Read error:", err)
			break
		}

		switch msg["type"] {
		case "register":
			currentMage.Username = msg["username"]
			currentMage.Password, err = HashPassword(msg["password"])
			if err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
				return
			}

			err := h.mageService.Create(currentMage)
			if err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
				continue
			}

			conn.WriteJSON(map[string]string{"type": "response", "message": "ok"})
		case "join":
			currentMage.Username = msg["username"]
			currentMage.Password = msg["password"]

			m, err := h.mageService.GetByUsername(currentMage.Username)
			if err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
				continue
			}

			if CheckPasswordHash(currentMage.Password, m.Password) {
				currentMage.ID = m.ID
				currentMage.HP = m.HP
				currentMage.Conn = conn

				if lastJoinedMage != nil {
					h.arenaService.RemoveMage(lastJoinedMage.Username)
					arenaMages := h.arenaService.GetMagesExcept(lastJoinedMage.Username)
					for _, v := range arenaMages {
						v.Conn.WriteJSON(map[string]interface{}{"type": "left", "username": currentMage.Username})
					}
				}

				h.arenaService.AddMage(currentMage)

				lastJoinedMage = &currentMage

				arenaMages := h.arenaService.GetMagesExcept(currentMage.Username)
				arenaMagesNames := make([]string, 0, len(arenaMages))
				for _, v := range arenaMages {
					v.Conn.WriteJSON(map[string]interface{}{"type": "joined", "username": currentMage.Username})
					arenaMagesNames = append(arenaMagesNames, v.Username)
				}

				conn.WriteJSON(map[string]interface{}{"type": "health", "hp": currentMage.HP, "mages": arenaMagesNames})
			} else {
				conn.WriteJSON(map[string]string{"type": "error", "message": "Invalid username/password"})
				continue
			}
		case "fireball":
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
			targetMage.Conn.WriteJSON(map[string]interface{}{"type": "attack", "from": currentMage.Username, "currentHP": targetMage.HP})

			if targetMage.HP <= 0 {
				targetMage.Conn.WriteJSON(map[string]interface{}{"type": "died", "by": currentMage.Username, "message": "busted"})

				h.arenaService.RemoveMage(targetMage.Username)
				arenaMages := h.arenaService.GetMagesExcept(targetMage.Username)
				for _, v := range arenaMages {
					v.Conn.WriteJSON(map[string]interface{}{"type": "died", "username": targetMage.Username})
				}

				targetMage.Conn.Close()
			}
		}
	}

	if lastJoinedMage == nil {
		return
	}

	h.arenaService.RemoveMage(currentMage.Username)
	arenaMages = h.arenaService.GetMagesExcept(currentMage.Username)
	for _, v := range arenaMages {
		v.Conn.WriteJSON(map[string]interface{}{"type": "left", "username": currentMage.Username})
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
