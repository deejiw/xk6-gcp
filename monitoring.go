package gcp

import (
	"context"
	"fmt"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func (g *Gcp) QueryTimeSeries(projectId string, query string) ([]*monitoringpb.TimeSeriesData, error) {
	ctx := context.Background()

	jwt, err := getJwtConfig(g.keyByte, g.scope)
	if err != nil {
		return nil, err
	}

	c, err := monitoring.NewQueryClient(ctx, option.WithTokenSource(jwt.TokenSource(ctx)))

	if err != nil {
		return nil, fmt.Errorf("Could not initialize query client <%w>", err)
	}

	defer c.Close()

	req := &monitoringpb.QueryTimeSeriesRequest{
		Name:  "projects/" + projectId,
		Query: query,
	}

	iter := c.QueryTimeSeries(ctx, req)

	var result []*monitoringpb.TimeSeriesData

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("Could not list time series: %w", err)
		}
		result = append(result, resp)
	}

	return result, nil
}
