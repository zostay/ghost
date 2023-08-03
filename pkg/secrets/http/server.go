package http

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/zostay/ghost/pkg/secrets"
)

type Server struct {
	UnimplementedKeeperServer
	secrets.Keeper
}

var _ KeeperServer = &Server{}

func NewServer(keeper secrets.Keeper) *Server {
	return &Server{
		Keeper: keeper,
	}
}

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
