package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"github.com/sirgwain/craig-stars/game"
)

// Allow a user to update a planet's orders
func (s *server) UpdatePlanetOrders(c *gin.Context) {
	user := s.GetSessionUser(c)

	var id idBind
	if err := c.ShouldBindUri(&id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// find the player for this user
	player, err := s.db.GetPlayerForGame(id.ID, user.ID)
	if err != nil {
		log.Error().Err(err).Int64("GameID", id.ID).Int64("UserID", user.ID).Msg("load planet from database")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load player from database"})
		return
	}

	planet := game.Planet{}
	if err := c.ShouldBindJSON(&planet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// find the existing planet by id
	existing, err := s.db.GetPlanet(planet.ID)
	if err != nil {
		log.Error().Err(err).Int64("ID", planet.ID).Msg("load planet from database")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to load planet from database"})
		return
	}

	// verify the user actually owns this planet
	if existing.PlayerNum != player.Num {
		c.JSON(http.StatusForbidden, gin.H{"error": fmt.Errorf("%s does not own %s", player, existing)})
		return
	}

	client := game.NewClient()
	client.UpdatePlanetOrders(player, existing, planet.ProductionQueue, planet.ContributesOnlyLeftoverToResearch)
	if err := s.db.UpdatePlanet(existing); err != nil {
		log.Error().Err(err).Int64("ID", existing.ID).Msg("update planet in database")
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to save planet to database"})
		return
	}

	c.JSON(http.StatusOK, existing)
}
