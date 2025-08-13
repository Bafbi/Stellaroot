package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/bafbi/stellaroot/libs/constant"
	"github.com/bafbi/stellaroot/libs/metadata"
	"github.com/bafbi/stellaroot/services/dashboard/templates"
)

type DashboardServer struct {
	metadataClient *metadata.Client
	logger         *slog.Logger
	router         *gin.Engine
}

// Use view models defined in the templates package for both JSON and fragments
type PlayerViewModel = templates.PlayerViewModel
type ServerViewModel = templates.ServerViewModel

func NewDashboardServer(metadataClient *metadata.Client, logger *slog.Logger) *DashboardServer {
	ds := &DashboardServer{
		metadataClient: metadataClient,
		logger:         logger,
	}

	ds.setupRouter()
	return ds
}

func (ds *DashboardServer) setupRouter() {
	ds.router = gin.Default()

	// Static files
	ds.router.Static("/static", "./static")

	// Routes using Templ
	ds.router.GET("/", ds.handleHome)
	ds.router.GET("/players", ds.handlePlayersPage)
	ds.router.GET("/players/fragment", ds.handlePlayersFragment)
	ds.router.GET("/servers", ds.handleServersPage)
	ds.router.GET("/servers/fragment", ds.handleServersFragment)

	// API routes
	api := ds.router.Group("/api")
	{
		api.GET("/players", ds.handlePlayersAPI)
		api.GET("/servers", ds.handleServersAPI)
		api.POST("/players/:uuid/update", ds.handleUpdatePlayer)
		api.POST("/servers/:name/update", ds.handleUpdateServer)
	}
}

func (ds *DashboardServer) handleHome(c *gin.Context) {
	playersCount := len(ds.metadataClient.GetPlayersByLabels(map[string]string{}))
	serversCount := len(ds.metadataClient.GetServersByLabels(map[string]string{}))

	component := templates.Index(playersCount, serversCount)
	component.Render(c.Request.Context(), c.Writer)
}

func (ds *DashboardServer) handlePlayersPage(c *gin.Context) {
	component := templates.Players()
	component.Render(c.Request.Context(), c.Writer)
}

func (ds *DashboardServer) handlePlayersFragment(c *gin.Context) {
	// Build view models from current cache
	players := ds.metadataClient.GetPlayersByLabels(map[string]string{})
	var viewModels []PlayerViewModel
	for uuid, player := range players {
		name := "Unknown"
		if player.Annotations != nil {
			if playerName, ok := player.Annotations[string(constant.PlayerName)]; ok {
				name = playerName
			}
		}
		status := "Offline"
		if player.Annotations != nil {
			if online, ok := player.Annotations["online"]; ok && online == "true" {
				status = "Online"
			}
		}
		viewModels = append(viewModels, PlayerViewModel{
			UUID:        uuid,
			Name:        name,
			Labels:      player.Labels,
			Annotations: player.Annotations,
			Status:      status,
		})
	}
	templates.PlayersFragment(viewModels).Render(c.Request.Context(), c.Writer)
}

func (ds *DashboardServer) handleServersPage(c *gin.Context) {
	component := templates.Servers()
	component.Render(c.Request.Context(), c.Writer)
}

func (ds *DashboardServer) handleServersFragment(c *gin.Context) {
	servers := ds.metadataClient.GetAllServers()
	var viewModels []ServerViewModel
	for name, server := range servers {
		status := "Unknown"
		playerCount := 0
		if server.Annotations != nil {
			if serverStatus, ok := server.Annotations["status"]; ok {
				status = serverStatus
			}
			if playerCountStr, ok := server.Annotations["current_players"]; ok {
				if count, err := strconv.Atoi(playerCountStr); err == nil {
					playerCount = count
				}
			}
		}
		viewModels = append(viewModels, ServerViewModel{
			Name:        name,
			Labels:      server.Labels,
			Annotations: server.Annotations,
			Status:      status,
			PlayerCount: playerCount,
		})
	}
	templates.ServersFragment(viewModels).Render(c.Request.Context(), c.Writer)
}

func (ds *DashboardServer) handlePlayersAPI(c *gin.Context) {
	players := ds.metadataClient.GetPlayersByLabels(map[string]string{})

	var viewModels []PlayerViewModel
	for uuid, player := range players {
		name := "Unknown"
		if player.Annotations != nil {
			if playerName, ok := player.Annotations[string(constant.PlayerName)]; ok {
				name = playerName
			}
		}

		status := "Offline"
		if player.Annotations != nil {
			if online, ok := player.Annotations["online"]; ok && online == "true" {
				status = "Online"
			}
		}

		viewModels = append(viewModels, PlayerViewModel{
			UUID:        uuid,
			Name:        name,
			Labels:      player.Labels,
			Annotations: player.Annotations,
			Status:      status,
		})
	}

	c.JSON(http.StatusOK, viewModels)
}

func (ds *DashboardServer) handleServersAPI(c *gin.Context) {
	servers := ds.metadataClient.GetAllServers()

	var viewModels []ServerViewModel
	for name, server := range servers {
		status := "Unknown"
		playerCount := 0

		if server.Annotations != nil {
			if serverStatus, ok := server.Annotations["status"]; ok {
				status = serverStatus
			}
			if playerCountStr, ok := server.Annotations["current_players"]; ok {
				if count, err := strconv.Atoi(playerCountStr); err == nil {
					playerCount = count
				}
			}
		}

		viewModels = append(viewModels, ServerViewModel{
			Name:        name,
			Labels:      server.Labels,
			Annotations: server.Annotations,
			Status:      status,
			PlayerCount: playerCount,
		})
	}

	c.JSON(http.StatusOK, viewModels)
}

func (ds *DashboardServer) handleUpdatePlayer(c *gin.Context) {
	uuid := c.Param("uuid")

	var updateData struct {
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := ds.metadataClient.UpdatePlayer(uuid, func(m *metadata.Metadata) {
		// Update labels
		for key, value := range updateData.Labels {
			if value == "" {
				m.DeleteLabel(key)
			} else {
				m.SetLabel(key, value)
			}
		}

		// Update annotations
		for key, value := range updateData.Annotations {
			if value == "" {
				m.DeleteAnnotation(key)
			} else {
				m.SetAnnotation(key, value)
			}
		}
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Player updated successfully"})
}

func (ds *DashboardServer) handleUpdateServer(c *gin.Context) {
	name := c.Param("name")

	var updateData struct {
		Labels      map[string]string `json:"labels"`
		Annotations map[string]string `json:"annotations"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := ds.metadataClient.UpdateServer(name, func(m *metadata.Metadata) {
		// Update labels
		for key, value := range updateData.Labels {
			if value == "" {
				m.DeleteLabel(key)
			} else {
				m.SetLabel(key, value)
			}
		}

		// Update annotations
		for key, value := range updateData.Annotations {
			if value == "" {
				m.DeleteAnnotation(key)
			} else {
				m.SetAnnotation(key, value)
			}
		}
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Server updated successfully"})
}

func (ds *DashboardServer) Start(addr string) error {
	ds.logger.Info("Starting dashboard server", "addr", addr)
	return ds.router.Run(addr)
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger.Info("Starting Stellaroot Dashboard")

	// Initialize metadata client
	config := metadata.NewConfigFromEnv()
	metadataClient, err := metadata.NewClient(context.Background(), config, logger.With("component", "metadata-client"))
	if err != nil {
		logger.Error("Failed to create metadata client", "error", err)
		os.Exit(1)
	}
	defer metadataClient.Close()

	// Create dashboard server
	dashboardServer := NewDashboardServer(metadataClient, logger)

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	if err := dashboardServer.Start(addr); err != nil {
		logger.Error("Failed to start dashboard server", "error", err)
		os.Exit(1)
	}
}
