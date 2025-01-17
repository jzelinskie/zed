package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	api "github.com/authzed/authzed-go/arrakisapi/api"
	"github.com/jzelinskie/cobrautil"
	"github.com/jzelinskie/stringz"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func SplitObject(s string) (nsName, id string, err error) {
	exploded := strings.Split(s, ":")
	if len(exploded) != 2 {
		return "", "", fmt.Errorf("invalid object format: %s", s)
	}
	return exploded[0], exploded[1], nil
}

func writeCmdFunc(operation api.RelationTupleUpdate_Operation) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		userNS, userID, err := SplitObject(args[0])
		if err != nil {
			return err
		}

		objectNS, objectID, err := SplitObject(args[1])
		if err != nil {
			return err
		}

		relation := args[2]

		tenant, token, endpoint, err := CurrentContext(cmd, contextConfigStore, tokenStore)
		if err != nil {
			return err
		}

		client, err := ClientFromFlags(cmd, endpoint, token)
		if err != nil {
			return err
		}

		request := &api.WriteRequest{Updates: []*api.RelationTupleUpdate{{
			Operation: operation,
			Tuple: &api.RelationTuple{
				ObjectAndRelation: &api.ObjectAndRelation{
					Namespace: stringz.Join("/", tenant, objectNS),
					ObjectId:  objectID,
					Relation:  relation,
				},
				User: &api.User{UserOneof: &api.User_Userset{Userset: &api.ObjectAndRelation{
					Namespace: stringz.Join("/", tenant, userNS),
					ObjectId:  userID,
					Relation:  "...",
				}}},
			},
		}}}

		resp, err := client.Write(context.Background(), request)
		if err != nil {
			return err
		}

		if cobrautil.MustGetBool(cmd, "json") || !term.IsTerminal(int(os.Stdout.Fd())) {
			prettyProto, err := prettyProto(resp)
			if err != nil {
				return err
			}

			fmt.Println(string(prettyProto))
			return nil
		}

		fmt.Println(resp.GetRevision().GetToken())

		return nil
	}
}
