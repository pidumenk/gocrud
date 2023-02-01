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

package database

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/RiskIdent/gocrud/pkg/model"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrBadRequest = errors.New("bad request")
)

type Client interface {
	CreateServer(ctx context.Context, server model.NewServer) (string, error)
	GetServer(ctx context.Context, id string) (model.Server, error)
	Close() error
}

type MongoDBClient struct {
	uri        string
	mongo      *mongo.Client
	serversCol *mongo.Collection
}

// Ensures it implements the interface
var _ Client = &MongoDBClient{}

func ConnectMongoDB(ctx context.Context, uri, db string) (Client, error) {
	log.Info().Str("mongouri", sanitizeURI(uri)).Msg("Connecting to mongodb.")
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, err
	}
	log.Debug().Msg("Connected to mongodb.")
	return &MongoDBClient{
		uri:        uri,
		mongo:      client,
		serversCol: client.Database(db).Collection("servers"),
	}, nil
}

func (c *MongoDBClient) CreateServer(ctx context.Context, server model.NewServer) (string, error) {
	res, err := c.serversCol.InsertOne(ctx, server)
	if err != nil {
		return "", err
	}
	id, ok := res.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("want primitive.ObjectID, got %T", res.InsertedID)
	}
	return id.Hex(), nil
}

func (c *MongoDBClient) GetServer(ctx context.Context, id string) (model.Server, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return model.Server{}, fmt.Errorf("%w, %s", ErrBadRequest, err)
	}
	var server model.Server
	if err := c.serversCol.
		FindOne(ctx, bson.D{{Key: "_id", Value: objectId}}).
		Decode(&server); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return model.Server{}, ErrNotFound
		}
		return model.Server{}, err
	}
	return server, nil
}

func (c *MongoDBClient) Close() error {
	if c.mongo == nil {
		return nil
	}
	log.Debug().Msg("Disconnecting from mongodb.")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err := c.mongo.Disconnect(ctx)
	c.mongo = nil
	return err
}

func sanitizeURI(value string) string {
	u, err := url.Parse(value)
	if err != nil {
		return "*censoring invalid url*"
	}
	if u.User != nil {
		u.User = url.UserPassword("...", "...")
	}
	return u.String()
}
