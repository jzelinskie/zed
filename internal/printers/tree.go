package printers

import (
	"fmt"

	api "github.com/authzed/authzed-go/arrakisapi/api"
	"github.com/cockroachdb/cockroach/pkg/util/treeprinter"
	"github.com/jzelinskie/stringz"
)

func prettyUser(user *api.User) string {
	userset := user.GetUserset()
	if userset.Relation == "..." {
		return fmt.Sprintf(
			"%s:%s",
			stringz.TrimPrefixIndex(userset.Namespace, "/"),
			userset.ObjectId,
		)
	}
	return fmt.Sprintf(
		"%s:%s %s",
		stringz.TrimPrefixIndex(userset.Namespace, "/"),
		userset.ObjectId,
		userset.Relation,
	)
}

func TreeNodeTree(tp treeprinter.Node, treeNode *api.RelationTupleTreeNode) {
	if treeNode.Expanded != nil {
		tp = tp.Child(fmt.Sprintf(
			"%s:%s %s",
			stringz.TrimPrefixIndex(treeNode.Expanded.Namespace, "/"),
			treeNode.Expanded.ObjectId,
			treeNode.Expanded.Relation,
		))
	}
	switch typed := treeNode.NodeType.(type) {
	case *api.RelationTupleTreeNode_IntermediateNode:
		switch typed.IntermediateNode.Operation {
		case api.SetOperationUserset_UNION:
			union := tp.Child("union")
			for _, child := range typed.IntermediateNode.ChildNodes {
				TreeNodeTree(union, child)
			}
		case api.SetOperationUserset_INTERSECTION:
			intersection := tp.Child("intersection")
			for _, child := range typed.IntermediateNode.ChildNodes {
				TreeNodeTree(intersection, child)
			}
		case api.SetOperationUserset_EXCLUSION:
			exclusion := tp.Child("exclusion")
			for _, child := range typed.IntermediateNode.ChildNodes {
				TreeNodeTree(exclusion, child)
			}
		default:
			panic("unknown expand operation")
		}
	case *api.RelationTupleTreeNode_LeafNode:
		for _, user := range typed.LeafNode.Users {
			tp.Child(prettyUser(user))
		}
	default:
		panic("unknown TreeNode type")
	}
}

func NamespaceTree(tp treeprinter.Node, nsdef *api.NamespaceDefinition) {
	root := tp.Child(stringz.TrimPrefixIndex(nsdef.GetName(), "/"))
	for _, relation := range nsdef.GetRelation() {
		relNode := root.Child(relation.GetName())
		if rewrite := relation.GetUsersetRewrite(); rewrite != nil {
			UsersetRewriteTree(relNode, rewrite)
		}
	}
}

func SetOperationChildTree(tp treeprinter.Node, child *api.SetOperation_Child) {
	switch child.ChildType.(type) {
	case *api.SetOperation_Child_XThis:
		tp.Child("_this")
	case *api.SetOperation_Child_ComputedUserset:
		ComputedUsersetTree(tp, child.GetComputedUserset())
	case *api.SetOperation_Child_TupleToUserset:
		TupleToUsersetTree(tp, child.GetTupleToUserset())
	case *api.SetOperation_Child_UsersetRewrite:
		UsersetRewriteTree(tp, child.GetUsersetRewrite())
	}
}

func ComputedUsersetTree(tp treeprinter.Node, userset *api.ComputedUserset) {
	obj := userset.GetObject()
	switch obj {
	case api.ComputedUserset_TUPLE_OBJECT:
		tp.Child("TUPLE_OBJECT: " + userset.GetRelation())
	case api.ComputedUserset_TUPLE_USERSET_OBJECT:
		tp.Child("TUPLE_USERSET_OBJECT " + userset.GetRelation())
	}
}

func TupleToUsersetTree(tp treeprinter.Node, t2u *api.TupleToUserset) {
	relNode := tp.Child(t2u.GetTupleset().GetRelation())
	ComputedUsersetTree(relNode, t2u.GetComputedUserset())
}

func UsersetRewriteTree(tp treeprinter.Node, rewrite *api.UsersetRewrite) {
	switch rewrite.GetRewriteOperation().(type) {
	case *api.UsersetRewrite_Union:
		childNode := tp.Child("union")
		for _, child := range rewrite.GetUnion().GetChild() {
			SetOperationChildTree(childNode, child)
		}
	case *api.UsersetRewrite_Intersection:
		childNode := tp.Child("intersection")
		for _, child := range rewrite.GetIntersection().GetChild() {
			SetOperationChildTree(childNode, child)
		}
	case *api.UsersetRewrite_Exclusion:
		childNode := tp.Child("exclusion")
		for _, child := range rewrite.GetExclusion().GetChild() {
			SetOperationChildTree(childNode, child)
		}
	}
}
