package main

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/cristalhq/acmd"
	"github.com/inenagl/anti-brute-force/internal/api"
	"google.golang.org/grpc"
)

type exitCmd struct {
	stop context.CancelFunc
}

func (c *exitCmd) ExecCommand(_ context.Context, _ []string) error {
	c.stop()
	return nil
}

func ExitCmd(stop context.CancelFunc) acmd.Command {
	ex := &exitCmd{stop: stop}
	return acmd.Command{
		Name:        "exit",
		Description: "exit",
		Exec:        ex,
	}
}

type authCmd struct {
	out    io.Writer
	client api.AntiBruteForceClient
}

func (c *authCmd) ExecCommand(ctx context.Context, args []string) error {
	if len(args) != 3 {
		return errors.New("wrong number of arguments, expected 3 arguments")
	}

	res, err := c.client.Auth(ctx, &api.AuthRequest{Login: args[0], Password: args[1], Ip: args[2]})
	if err != nil {
		return err
	}
	if _, err = fmt.Fprintf(c.out, "Response: Ok: %v\n", res.GetOk()); err != nil {
		return err
	}

	return nil
}

func AuthCmd(out io.Writer, client api.AntiBruteForceClient) acmd.Command {
	ex := &authCmd{out: out, client: client}
	return acmd.Command{
		Name:        "auth",
		Description: "check auth request to anti-bruteforce service",
		Exec:        ex,
	}
}

type resetCmd struct {
	out    io.Writer
	client api.AntiBruteForceClient
}

func (c *resetCmd) ExecCommand(ctx context.Context, args []string) error {
	if len(args) != 3 {
		return errors.New("number of arguments not equal 3")
	}

	_, err := c.client.Reset(ctx, &api.ResetRequest{Login: args[0], Password: args[1], Ip: args[2]})
	if err != nil {
		if _, err = fmt.Fprintf(c.out, "Error: %v\n", err); err != nil {
			return err
		}
	}
	if _, err = fmt.Fprint(c.out, "Response: Ok\n"); err != nil {
		return err
	}

	return nil
}

func ResetCmd(out io.Writer, client api.AntiBruteForceClient) acmd.Command {
	ex := &resetCmd{out: out, client: client}
	return acmd.Command{
		Name:        "reset",
		Description: "reset buckets for given login, password and IP",
		Exec:        ex,
	}
}

type bwListOperationCfg struct {
	out    io.Writer
	client api.AntiBruteForceClient
}

type addToWhiteCmd struct {
	cfg bwListOperationCfg
}

type addToBlackCmd struct {
	cfg bwListOperationCfg
}

type removeFromWhiteCmd struct {
	cfg bwListOperationCfg
}

type removeFromBlackCmd struct {
	cfg bwListOperationCfg
}

func (c *addToWhiteCmd) ExecCommand(ctx context.Context, args []string) error {
	return processBWRequest(ctx, args, c.cfg.out, c.cfg.client.AddToWhiteList)
}

func (c *addToBlackCmd) ExecCommand(ctx context.Context, args []string) error {
	return processBWRequest(ctx, args, c.cfg.out, c.cfg.client.AddToBlackList)
}

func (c *removeFromWhiteCmd) ExecCommand(ctx context.Context, args []string) error {
	return processBWRequest(ctx, args, c.cfg.out, c.cfg.client.RemoveFromWhiteList)
}

func (c *removeFromBlackCmd) ExecCommand(ctx context.Context, args []string) error {
	return processBWRequest(ctx, args, c.cfg.out, c.cfg.client.RemoveFromBlackList)
}

func processBWRequest(ctx context.Context, args []string, out io.Writer,
	exec func(c context.Context, req *api.IpNetRequest, opts ...grpc.CallOption) (*api.EmptyResponse, error),
) error {
	if len(args) != 1 {
		return errors.New("ipNet argument is missing")
	}

	_, err := exec(ctx, &api.IpNetRequest{Inet: args[0]})
	if err != nil {
		return err
	}
	if _, err = fmt.Fprint(out, "Response: Ok\n"); err != nil {
		return err
	}

	return nil
}

func AddCmd(out io.Writer, client api.AntiBruteForceClient) acmd.Command {
	exWhite := &addToWhiteCmd{cfg: bwListOperationCfg{out: out, client: client}}
	exBlack := &addToBlackCmd{cfg: bwListOperationCfg{out: out, client: client}}
	return acmd.Command{
		Name:        "add",
		Description: "adds IpNet to black/white list",
		Subcommands: []acmd.Command{
			{
				Name:        "white",
				Description: "adds IpNet to white list",
				Exec:        exWhite,
			},
			{
				Name:        "black",
				Description: "adds IpNet to black list",
				Exec:        exBlack,
			},
		},
	}
}

func RemoveCmd(out io.Writer, client api.AntiBruteForceClient) acmd.Command {
	exWhite := &removeFromWhiteCmd{cfg: bwListOperationCfg{out: out, client: client}}
	exBlack := &removeFromBlackCmd{cfg: bwListOperationCfg{out: out, client: client}}
	return acmd.Command{
		Name:        "remove",
		Description: "removes IpNet from black/white list",
		Subcommands: []acmd.Command{
			{
				Name:        "white",
				Description: "remove IpNet from white list",
				Exec:        exWhite,
			},
			{
				Name:        "black",
				Description: "remove IpNet from black list",
				Exec:        exBlack,
			},
		},
	}
}
