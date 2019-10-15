package waitinglist

import (
	"errors"
	"fmt"
	"github.com/briscola-as-a-service/game"
	"github.com/briscola-as-a-service/game/player"
	"sync"
)

var mux sync.Mutex

type WaitingLists map[string]*WaitingList

var waitingLists WaitingLists

type WaitingList struct {
	name          string
	playersNumber int
	players       []player.Player
}

func init() {
	waitingLists = make(map[string]*WaitingList)
}

func New() *WaitingLists {
	return &waitingLists
}

func (wls *WaitingLists) AddList(listName string, playersNumber int) error {
	if _, exists := (*wls)[listName]; exists == true {
		return errors.New("Waiting List Exists")
	}
	players := make([]player.Player, 0)
	wl := WaitingList{listName, playersNumber, players}
	(*wls)[listName] = &wl
	return nil
}

func (wls *WaitingLists) AddPlayer(listName string, playerName string, playerID string) error {
	// Protect from new player race conditions! No new players can be added until
	// is verified if a new game can start (StartGame())
	mux.Lock()

	if _, exists := (*wls)[listName]; exists == false {
		return errors.New("Waiting list does not exists")
	}
	if len((*wls)[listName].players) > (*wls)[listName].playersNumber-1 {
		return errors.New("Too many players")
	}
	player := player.New(playerName, playerID)

	waitingListPtr := (*wls)[listName]
	players := (*waitingListPtr).players

	for _, p := range players {
		if p.Is(player) {
			return errors.New("Player is already in waiting list")
		}
	}

	players = append(players, player)
	(*waitingListPtr).players = players

	return nil
}

func (wls *WaitingLists) StartGame(listName string) (decker *game.Decker, err error) {
	// Protect from new player race conditions! No new players can be added until
	// mux created in AddPlayer
	defer mux.Unlock()

	if _, exists := (*wls)[listName]; exists == false {
		err = errors.New("Waiting list does not exists")
		return
	}
	if len((*wls)[listName].players) < (*wls)[listName].playersNumber {
		err = errors.New("Waiting for players")
		return
	}

	waitingListPtr := (*wls)[listName]
	players := (*waitingListPtr).players

	d := game.New(players)
	decker = &d

	// Reset the waiting list
	emptyPlayers := make([]player.Player, 0)
	(*waitingListPtr).players = emptyPlayers

	return
}

func show(i interface{}) {
	fmt.Printf("WL:  %+v\n", i)
}
