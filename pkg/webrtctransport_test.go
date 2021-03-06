package sfu

import (
	"context"
	"reflect"
	"testing"

	"github.com/pion/sdp/v3"

	"github.com/pion/webrtc/v3"
	"github.com/stretchr/testify/assert"
)

func TestNewWebRTCTransport(t *testing.T) {
	type args struct {
		ctx     context.Context
		session *Session
		me      MediaEngine
		cfg     WebRTCTransportConfig
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Must create a non nil webRTC transport",
			args: args{
				ctx:     context.Background(),
				session: NewSession("test"),
				me:      MediaEngine{},
				cfg:     WebRTCTransportConfig{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewWebRTCTransport(tt.args.session, tt.args.me, tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewWebRTCTransport() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NotNil(t, got)
		})
	}
}

func TestWebRTCTransport_Close(t *testing.T) {
	me := webrtc.MediaEngine{}
	me.RegisterDefaultCodecs()
	api := webrtc.NewAPI(webrtc.WithMediaEngine(me))
	peer, err := api.NewPeerConnection(webrtc.Configuration{})
	assert.NoError(t, err)
	s := NewSession("session")

	type fields struct {
		id      string
		pc      *webrtc.PeerConnection
		session *Session
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Must close inner peer connection and cancel context without errors.",
			fields: fields{
				id:      "test",
				pc:      peer,
				session: s,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := &WebRTCTransport{
				id:      tt.fields.id,
				pc:      tt.fields.pc,
				session: tt.fields.session,
			}
			s.transports[p.ID()] = p
			if err := p.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
			assert.Equal(t, 0, len(s.transports))
		})
	}
}

func TestWebRTCTransport_CreateAnswer(t *testing.T) {
	me := webrtc.MediaEngine{}
	me.RegisterDefaultCodecs()
	api := webrtc.NewAPI(webrtc.WithMediaEngine(me))
	sfu, remote, err := newPair(webrtc.Configuration{}, api)
	assert.NoError(t, err)

	remoteTrack, err := remote.NewTrack(webrtc.DefaultPayloadTypeVP8, 1234, "video", "pion")
	assert.NoError(t, err)
	_, err = remote.AddTrack(remoteTrack)
	assert.NoError(t, err)

	offer, err := remote.CreateOffer(nil)
	assert.NoError(t, err)
	gatherComplete := webrtc.GatheringCompletePromise(remote)
	err = remote.SetLocalDescription(offer)
	assert.NoError(t, err)
	<-gatherComplete
	err = sfu.SetRemoteDescription(offer)
	assert.NoError(t, err)

	type fields struct {
		pc *webrtc.PeerConnection
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Must return answer without errors",
			fields: fields{
				pc: sfu,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := &WebRTCTransport{
				pc: tt.fields.pc,
			}
			_, err := p.CreateAnswer()
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateAnswer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestWebRTCTransport_CreateOffer(t *testing.T) {
	me := webrtc.MediaEngine{}
	me.RegisterDefaultCodecs()
	api := webrtc.NewAPI(webrtc.WithMediaEngine(me))
	sfu, err := api.NewPeerConnection(webrtc.Configuration{})
	assert.NoError(t, err)

	remoteTrack, err := sfu.NewTrack(webrtc.DefaultPayloadTypeVP8, 1234, "video", "pion")
	assert.NoError(t, err)
	_, err = sfu.AddTrack(remoteTrack)
	assert.NoError(t, err)

	type fields struct {
		pc *webrtc.PeerConnection
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "Must return offer without errors",
			fields:  fields{pc: sfu},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := &WebRTCTransport{
				pc: tt.fields.pc,
			}
			_, err := p.CreateOffer()
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateOffer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestWebRTCTransport_GetRouter(t *testing.T) {
	type fields struct {
		router Router
	}
	type args struct {
		trackID string
	}

	router := &router{}

	tests := []struct {
		name   string
		fields fields
		args   args
		want   Router
	}{
		{
			name: "Must return router by ID",
			fields: fields{
				router: router,
			},
			args: args{
				trackID: "test",
			},
			want: router,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := &WebRTCTransport{
				router: tt.fields.router,
			}
			if got := p.GetRouter(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRouter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWebRTCTransport_ID(t *testing.T) {
	type fields struct {
		id string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Must return current ID",
			fields: fields{
				id: "test",
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := &WebRTCTransport{
				id: tt.fields.id,
			}
			if got := p.ID(); got != tt.want {
				t.Errorf("ID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWebRTCTransport_LocalDescription(t *testing.T) {
	me := webrtc.MediaEngine{}
	me.RegisterDefaultCodecs()
	api := webrtc.NewAPI(webrtc.WithMediaEngine(me))
	sfu, err := api.NewPeerConnection(webrtc.Configuration{})
	assert.NoError(t, err)

	remoteTrack, err := sfu.NewTrack(webrtc.DefaultPayloadTypeVP8, 1234, "video", "pion")
	assert.NoError(t, err)
	_, err = sfu.AddTrack(remoteTrack)
	assert.NoError(t, err)

	offer, err := sfu.CreateOffer(nil)
	assert.NoError(t, err)
	gatherComplete := webrtc.GatheringCompletePromise(sfu)
	err = sfu.SetLocalDescription(offer)
	assert.NoError(t, err)
	<-gatherComplete
	targetLD := sfu.LocalDescription()

	type fields struct {
		pc *webrtc.PeerConnection
	}
	tests := []struct {
		name   string
		fields fields
		want   *webrtc.SessionDescription
	}{
		{
			name: "Must return current peer local description",
			fields: fields{
				pc: sfu,
			},
			want: targetLD,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := &WebRTCTransport{
				pc: tt.fields.pc,
			}
			if got := p.LocalDescription(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LocalDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWebRTCTransport_OnTrack(t *testing.T) {
	type args struct {
		f func(*webrtc.Track, *webrtc.RTPReceiver)
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Must set on track func",
			args: args{
				f: func(_ *webrtc.Track, _ *webrtc.RTPReceiver) {
				},
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := &WebRTCTransport{}
			p.OnTrack(tt.args.f)
			assert.NotNil(t, p.onTrackHandler)
		})
	}
}

func TestWebRTCTransport_SetLocalDescription(t *testing.T) {
	me := webrtc.MediaEngine{}
	me.RegisterDefaultCodecs()
	api := webrtc.NewAPI(webrtc.WithMediaEngine(me))
	sfu, err := api.NewPeerConnection(webrtc.Configuration{})
	assert.NoError(t, err)
	offer, err := sfu.CreateOffer(nil)
	assert.NoError(t, err)

	type fields struct {
		pc *webrtc.PeerConnection
	}
	type args struct {
		desc webrtc.SessionDescription
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Must set local description on peer",
			fields: fields{
				pc: sfu,
			},
			args:    args{desc: offer},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := &WebRTCTransport{
				pc: tt.fields.pc,
			}
			if err := p.SetLocalDescription(tt.args.desc); (err != nil) != tt.wantErr {
				t.Errorf("SetLocalDescription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWebRTCTransport_SetRemoteDescription(t *testing.T) {
	me := webrtc.MediaEngine{}
	me.RegisterDefaultCodecs()
	api := webrtc.NewAPI(webrtc.WithMediaEngine(me))
	sfu, remote, err := newPair(webrtc.Configuration{}, api)
	assert.NoError(t, err)
	remoteTrack, err := sfu.NewTrack(webrtc.DefaultPayloadTypeVP8, 1234, "video", "pion")
	assert.NoError(t, err)
	_, err = sfu.AddTrack(remoteTrack)
	assert.NoError(t, err)

	offer, err := sfu.CreateOffer(nil)
	assert.NoError(t, err)
	gatherComplete := webrtc.GatheringCompletePromise(sfu)
	err = sfu.SetLocalDescription(offer)
	assert.NoError(t, err)
	<-gatherComplete
	err = remote.SetRemoteDescription(*sfu.LocalDescription())
	assert.NoError(t, err)
	answer, err := remote.CreateAnswer(nil)
	assert.NoError(t, err)
	err = remote.SetLocalDescription(answer)
	assert.NoError(t, err)

	type fields struct {
		pc *webrtc.PeerConnection
	}
	type args struct {
		desc webrtc.SessionDescription
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Must set remote description without errors",
			fields: fields{
				pc: sfu,
			},
			args: args{
				desc: answer,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := &WebRTCTransport{
				pc: tt.fields.pc,
				router: &RouterMock{
					SetExtMapFunc: func(_ *sdp.SessionDescription) {
					},
				},
			}
			if err := p.SetRemoteDescription(tt.args.desc); (err != nil) != tt.wantErr {
				t.Errorf("SetRemoteDescription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWebRTCTransport_AddSender(t *testing.T) {
	type fields struct {
		senders map[string][]Sender
	}
	type args struct {
		streamID string
		sender   Sender
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "Must add sender to given stream ID",
			fields: fields{senders: map[string][]Sender{}},
			args: struct {
				streamID string
				sender   Sender
			}{streamID: "test", sender: &SimpleSender{}},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := &WebRTCTransport{
				senders: tt.fields.senders,
			}
			p.AddSender(tt.args.streamID, tt.args.sender)
			assert.Equal(t, 1, len(p.senders))
			assert.Equal(t, 1, len(p.senders[tt.args.streamID]))
		})
	}
}

func TestWebRTCTransport_GetSenders(t *testing.T) {
	type fields struct {
		senders map[string][]Sender
	}
	type args struct {
		streamID string
	}
	sdrs := map[string][]Sender{"test": {&SimpleSender{}, &SimulcastSender{}}}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []Sender
	}{
		{
			name: "Must return an array of senders from given stream ID",
			fields: fields{
				senders: sdrs,
			},
			args: args{
				streamID: "test",
			},
			want: sdrs["test"],
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			p := &WebRTCTransport{
				senders: tt.fields.senders,
			}
			if got := p.GetSenders(tt.args.streamID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSenders() = %v, want %v", got, tt.want)
			}
		})
	}
}
