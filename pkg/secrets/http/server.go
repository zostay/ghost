package http

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/zostay/ghost/pkg/secrets"
)

// Server gives a secret keeper a gRPC server interface.
type Server struct {
	UnimplementedKeeperServer
	secrets.Keeper

	name              string
	enforcementPeriod time.Duration
	enforcedPolicies  []string
}

var _ KeeperServer = &Server{}

// NewServer creates a new gRPC server for the wrapped secret keeper.
func NewServer(
	keeper secrets.Keeper,
	name string,
	enforcementPeriod time.Duration,
	enforcedPolicies []string,
) *Server {
	return &Server{
		Keeper:            keeper,
		name:              name,
		enforcementPeriod: enforcementPeriod,
		enforcedPolicies:  enforcedPolicies,
	}
}

// ListLocations maps the ListLocations secret keeper call to the gRPC
// interface.
func (s *Server) ListLocations(
	_ *empty.Empty,
	stream Keeper_ListLocationsServer,
) error {
	locs, err := s.Keeper.ListLocations(stream.Context())
	if err != nil {
		return err
	}

	for _, loc := range locs {
		err := stream.Send(&Location{
			Location: loc,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// ListSecrets maps the ListSecrets secret keeper call to the gRPC interface.
func (s *Server) ListSecrets(
	location *Location,
	stream Keeper_ListSecretsServer,
) error {
	ids, err := s.Keeper.ListSecrets(stream.Context(), location.GetLocation())
	if err != nil {
		return err
	}

	for _, id := range ids {
		sec := &Secret{Id: id}
		err := stream.Send(sec)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetSecretsByName maps the GetSecretsByName secret keeper call to the gRPC
// interface.
func (s *Server) GetSecretsByName(
	req *GetSecretsByNameRequest,
	stream Keeper_GetSecretsByNameServer,
) error {
	secs, err := s.Keeper.GetSecretsByName(stream.Context(), req.GetName())
	if err != nil {
		return err
	}

	for _, sec := range secs {
		rpcSec := FromSecret(sec)
		err := stream.Send(rpcSec)
		if err != nil {
			return err
		}
	}

	return nil
}

// GetSecret maps the GetSecret secret keeper call to the gRPC interface.
func (s *Server) GetSecret(
	ctx context.Context,
	req *GetSecretRequest,
) (*Secret, error) {
	sec, err := s.Keeper.GetSecret(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return FromSecret(sec), nil
}

// SetSecret maps the SetSecret secret keeper call to the gRPC interface.
func (s *Server) SetSecret(
	ctx context.Context,
	rpcSec *Secret,
) (*Secret, error) {
	sec, err := s.Keeper.SetSecret(ctx, NewSecretWrapper(rpcSec))
	if err != nil {
		return nil, err
	}

	return FromSecret(sec), nil
}

// CopySecret maps the CopySecret secret keeper call to the gRPC interface.
func (s *Server) CopySecret(
	ctx context.Context,
	req *ChangeLocationRequest,
) (*Secret, error) {
	sec, err := s.Keeper.CopySecret(ctx, req.GetId(), req.GetLocation())
	if err != nil {
		return nil, err
	}

	return FromSecret(sec), nil
}

// MoveSecret maps the MoveSecret secret keeper call to the gRPC interface.
func (s *Server) MoveSecret(
	ctx context.Context,
	req *ChangeLocationRequest,
) (*Secret, error) {
	sec, err := s.Keeper.MoveSecret(ctx, req.GetId(), req.GetLocation())
	if err != nil {
		return nil, err
	}

	return FromSecret(sec), nil
}

// DeleteSecret maps the DeleteSecret secret keeper call to the gRPC interface.
func (s *Server) DeleteSecret(
	ctx context.Context,
	req *DeleteSecretRequest,
) (*empty.Empty, error) {
	err := s.Keeper.DeleteSecret(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// GetServiceInfo returns the service info for this server.
func (s *Server) GetServiceInfo(
	_ context.Context,
	_ *empty.Empty,
) (*ServiceInfo, error) {
	return &ServiceInfo{
		Keeper:            s.name,
		EnforcementPeriod: durationpb.New(s.enforcementPeriod),
		EnforcedPolicies:  s.enforcedPolicies,
	}, nil
}
