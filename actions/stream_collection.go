package actions

import "github.com/taiyoh/toyhose/datatypes/firehose"

type streamCollection []*firehose.DeliveryStream

func (c streamCollection) Names() []string {
	names := make([]string, 0, len(c))
	for _, stream := range c {
		names = append(names, stream.StreamName())
	}
	return names
}
