package api

import (
	"context"
	"fmt"
	"net"

	"github.com/inenagl/anti-brute-force/internal/app"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Service struct {
	UnimplementedAntiBruteForceServer
	app app.Application
	log *zap.Logger
}

func NewService(a app.Application, l *zap.Logger) *Service {
	return &Service{
		app: a,
		log: l,
	}
}

func (s *Service) Auth(_ context.Context, request *AuthRequest) (*AuthResponse, error) {
	login := request.GetLogin()
	passwd := request.GetPassword()
	ip := net.ParseIP(request.GetIp())
	if ip == nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", fmt.Sprintf(`"%s" is not valid ip`, request.GetIp()))
	}

	res, err := s.app.Auth(login, passwd, ip)
	if err != nil {
		s.log.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &AuthResponse{Ok: res}, nil
}

func (s *Service) Reset(_ context.Context, request *ResetRequest) (*EmptyResponse, error) {
	login := request.GetLogin()
	passwd := request.GetPassword()
	var ip net.IP
	if request.GetIp() != "" {
		ip := net.ParseIP(request.Ip)
		if ip == nil {
			return nil, status.Errorf(codes.InvalidArgument, "%s", fmt.Sprintf(`"%s" is not valid ip`, request.Ip))
		}
	}
	s.app.ResetBuckets(login, passwd, ip)

	return &EmptyResponse{}, nil
}

func parseIPNet(s string) (*net.IPNet, error) {
	_, IPNet, err := net.ParseCIDR(s)
	// Допускаются единичные IP, поэтому пробуем распарсить как IP
	if err != nil {
		ip := net.ParseIP(s)
		if ip == nil {
			return nil, fmt.Errorf(`%w: can't parse "%s" to IP Network`, err, s)
		}
		IPNet = &net.IPNet{
			IP:   ip,
			Mask: net.IPv4Mask(255, 255, 255, 255),
		}
	}
	return IPNet, nil
}

func (s *Service) AddToBlackList(_ context.Context, request *IpNetRequest) (*EmptyResponse, error) {
	network, err := parseIPNet(request.GetInet())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	if err = s.app.AddToBlackList(*network); err != nil {
		s.log.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &EmptyResponse{}, nil
}

func (s *Service) AddToWhiteList(_ context.Context, request *IpNetRequest) (*EmptyResponse, error) {
	network, err := parseIPNet(request.GetInet())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	if err = s.app.AddToWhiteList(*network); err != nil {
		s.log.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &EmptyResponse{}, nil
}

func (s *Service) RemoveFromBlackList(_ context.Context, request *IpNetRequest) (*EmptyResponse, error) {
	network, err := parseIPNet(request.GetInet())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	if err = s.app.RemoveFromBlackList(*network); err != nil {
		s.log.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &EmptyResponse{}, nil
}

func (s *Service) RemoveFromWhiteList(_ context.Context, request *IpNetRequest) (*EmptyResponse, error) {
	network, err := parseIPNet(request.GetInet())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%s", err.Error())
	}

	if err = s.app.RemoveFromWhiteList(*network); err != nil {
		s.log.Error(err.Error())
		return nil, status.Errorf(codes.Internal, "%s", err.Error())
	}

	return &EmptyResponse{}, nil
}

func (s *Service) mustEmbedUnimplementedAntiBruteForceServer() {
	// TODO implement me
	panic("implement me")
}
