package httpx

import (
	"testing"

	npanv1 "npan/gen/go/npan/v1"

	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestConnectTimestampDescriptor_ProgressMessagesHaveSidecarFields(t *testing.T) {
	t.Parallel()

	file := npanv1.File_npan_v1_api_proto
	messages := file.Messages()

	cases := []struct {
		messageName string
		fieldNames  []string
	}{
		{
			messageName: "CrawlStats",
			fieldNames:  []string{"started_at_ts", "ended_at_ts"},
		},
		{
			messageName: "RootSyncProgress",
			fieldNames:  []string{"updated_at_ts"},
		},
		{
			messageName: "SyncProgressState",
			fieldNames:  []string{"started_at_ts", "updated_at_ts"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.messageName, func(t *testing.T) {
			t.Parallel()

			msg := messages.ByName(protoreflect.Name(tc.messageName))
			if msg == nil {
				t.Fatalf("message %s not found", tc.messageName)
			}
			for _, fieldName := range tc.fieldNames {
				field := msg.Fields().ByName(protoreflect.Name(fieldName))
				if field == nil {
					t.Fatalf("expected field %s on message %s", fieldName, tc.messageName)
				}
				if field.Kind() != protoreflect.MessageKind {
					t.Fatalf("expected %s.%s to be message kind, got %s", tc.messageName, fieldName, field.Kind())
				}
			}
		})
	}
}
