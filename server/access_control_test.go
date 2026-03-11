package server

import (
	"context"
	"testing"

	"github.com/otfabric/opcua/id"
	"github.com/otfabric/opcua/ua"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// denyAllController denies all operations.
type denyAllController struct{}

func (denyAllController) CheckRead(_ context.Context, _ *session, _ *ua.NodeID) ua.StatusCode {
	return ua.StatusBadUserAccessDenied
}
func (denyAllController) CheckWrite(_ context.Context, _ *session, _ *ua.NodeID) ua.StatusCode {
	return ua.StatusBadUserAccessDenied
}
func (denyAllController) CheckBrowse(_ context.Context, _ *session, _ *ua.NodeID) ua.StatusCode {
	return ua.StatusBadUserAccessDenied
}
func (denyAllController) CheckCall(_ context.Context, _ *session, _ *ua.NodeID) ua.StatusCode {
	return ua.StatusBadUserAccessDenied
}

func TestAccessControl_ReadDenied(t *testing.T) {
	srv := newTestServer()
	srv.cfg.accessController = denyAllController{}
	ns, _ := addTestNamespace(srv)
	svc := &AttributeService{srv: srv}

	req := &ua.ReadRequest{
		RequestHeader: reqHeader(),
		NodesToRead: []*ua.ReadValueID{
			{NodeID: ua.NewStringNodeID(ns.ID(), "rw_int32"), AttributeID: ua.AttributeIDValue},
		},
	}
	resp, err := svc.Read(context.Background(), nil, req, 1)
	require.NoError(t, err)

	readResp := resp.(*ua.ReadResponse)
	require.Len(t, readResp.Results, 1)
	assert.Equal(t, ua.StatusBadUserAccessDenied, readResp.Results[0].Status)
}

func TestAccessControl_WriteDenied(t *testing.T) {
	srv := newTestServer()
	srv.cfg.accessController = denyAllController{}
	ns, _ := addTestNamespace(srv)
	svc := &AttributeService{srv: srv}

	req := &ua.WriteRequest{
		RequestHeader: reqHeader(),
		NodesToWrite: []*ua.WriteValue{
			{
				NodeID:      ua.NewStringNodeID(ns.ID(), "rw_int32"),
				AttributeID: ua.AttributeIDValue,
				Value: &ua.DataValue{
					EncodingMask: ua.DataValueValue,
					Value:        ua.MustVariant(int32(99)),
				},
			},
		},
	}
	resp, err := svc.Write(context.Background(), nil, req, 1)
	require.NoError(t, err)

	writeResp := resp.(*ua.WriteResponse)
	require.Len(t, writeResp.Results, 1)
	assert.Equal(t, ua.StatusBadUserAccessDenied, writeResp.Results[0])
}

func TestAccessControl_BrowseDenied(t *testing.T) {
	srv := newTestServer()
	srv.cfg.accessController = denyAllController{}
	ns, _ := addTestNamespace(srv)
	svc := &ViewService{srv: srv, cps: make(map[string]*continuationPoint)}

	req := &ua.BrowseRequest{
		RequestHeader: reqHeader(),
		NodesToBrowse: []*ua.BrowseDescription{{
			NodeID:          ua.NewNumericNodeID(ns.ID(), id.ObjectsFolder),
			BrowseDirection: ua.BrowseDirectionForward,
			ReferenceTypeID: ua.NewNumericNodeID(0, 0),
			IncludeSubtypes: true,
			ResultMask:      uint32(ua.BrowseResultMaskAll),
		}},
	}
	resp, err := svc.Browse(context.Background(), nil, req, 1)
	require.NoError(t, err)

	browseResp := resp.(*ua.BrowseResponse)
	require.Len(t, browseResp.Results, 1)
	assert.Equal(t, ua.StatusBadUserAccessDenied, browseResp.Results[0].StatusCode)
}

func TestAccessControl_CallDenied(t *testing.T) {
	srv := newTestServer()
	srv.cfg.accessController = denyAllController{}
	ns, obj := addTestNamespace(srv)
	svc := &MethodService{srv: srv}

	methodID := ua.NewStringNodeID(ns.ID(), "ac_method")
	methodNode := NewFolderNode(methodID, "ac_method")
	methodNode.SetNodeClass(ua.NodeClassMethod)
	ns.AddNode(methodNode)
	obj.AddRef(methodNode, RefType(id.HasComponent), true)

	srv.RegisterMethod(obj.ID(), methodID, func(ctx context.Context, oID, mID *ua.NodeID, args []*ua.Variant) ([]*ua.Variant, ua.StatusCode) {
		return nil, ua.StatusOK
	})

	req := &ua.CallRequest{
		RequestHeader: reqHeader(),
		MethodsToCall: []*ua.CallMethodRequest{{
			ObjectID: obj.ID(),
			MethodID: methodID,
		}},
	}
	resp, err := svc.Call(context.Background(), nil, req, 1)
	require.NoError(t, err)

	callResp := resp.(*ua.CallResponse)
	require.Len(t, callResp.Results, 1)
	assert.Equal(t, ua.StatusBadUserAccessDenied, callResp.Results[0].StatusCode)
}

func TestAccessControl_DefaultAllows(t *testing.T) {
	ac := DefaultAccessController{}
	assert.Equal(t, ua.StatusOK, ac.CheckRead(context.Background(), nil, nil))
	assert.Equal(t, ua.StatusOK, ac.CheckWrite(context.Background(), nil, nil))
	assert.Equal(t, ua.StatusOK, ac.CheckBrowse(context.Background(), nil, nil))
	assert.Equal(t, ua.StatusOK, ac.CheckCall(context.Background(), nil, nil))
}

func TestAccessControl_WithAccessControllerOption(t *testing.T) {
	srv := New(EndPoint("localhost", 4840), WithAccessController(denyAllController{}))
	assert.IsType(t, denyAllController{}, srv.cfg.accessController)
}
