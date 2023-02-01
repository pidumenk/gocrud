// SPDX-FileCopyrightText: 2022 Risk.Ident GmbH <contact@riskident.com>
//
// SPDX-License-Identifier: GPL-3.0-or-later
//
// This program is free software: you can redistribute it and/or modify it
// under the terms of the GNU General Public License as published by the
// Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT
// ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
// FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License for
// more details.
//
// You should have received a copy of the GNU General Public License along
// with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/RiskIdent/gocrud/pkg/database"
	"github.com/RiskIdent/gocrud/pkg/model"
	"github.com/alecthomas/kong"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var cli struct {
	BindAddress string `default:"0.0.0.0:8080" help:"Address to serve API on" env:"GOCRUD_BIND_ADDRESS"`
	MongoURI    string `default:"mongodb://localhost:27017" help:"MongoDB URI to use" env:"GOCRUD_MONGO_URI"`
	MongoDB     string `default:"gocrud" help:"MongoDB database to use" env:"GOCRUD_MONGO_DB"`
}

func main() {
	kong.Parse(&cli)

	initLogger()

	if err := mainE(); err != nil {
		log.Error().Err(err).Msg("Failed execution.")
		os.Exit(1)
	}
}

func mainE() error {
	db, err := database.ConnectMongoDB(context.Background(), cli.MongoURI, cli.MongoDB)
	if err != nil {
		return err
	}
	defer db.Close()

	gin.DefaultErrorWriter = log.Logger
	gin.DefaultWriter = log.Logger

	router := gin.New()

	router.Use(
		gin.LoggerWithConfig(gin.LoggerConfig{
			SkipPaths: []string{"/"},
		}),
		gin.Recovery(),
	)

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "Hello from gocrud :)"})
	})

	v1 := router.Group("/v1")
	{
		v1.POST("/server", func(c *gin.Context) {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()
			var server model.NewServer
			if err := c.ShouldBindJSON(&server); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			id, err := db.CreateServer(ctx, server)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			log.Info().Str("id", id).Msg("Created new server.")
			c.JSON(http.StatusOK, gin.H{"id": id})
		})

		v1.GET("/server/:id", func(c *gin.Context) {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()
			id := c.Param("id")
			server, err := db.GetServer(ctx, id)
			if errors.Is(err, database.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			if errors.Is(err, database.ErrBadRequest) {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, server)
		})
	}

	return router.Run(cli.BindAddress)
}

func initLogger() error {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "Jan-02 15:04",
	})
	return nil
}
