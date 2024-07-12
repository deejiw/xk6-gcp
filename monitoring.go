package gcp

import (
	"context"
	"fmt"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"golang.org/x/oauth2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// This function is querying time series data from Google Cloud Monitoring API. It takes in a project
// ID and a query string as parameters, and returns a slice of `monitoringpb.TimeSeriesData` and an
// error.
func (g *Gcp) QueryTimeSeries(projectId string, query string) ([]*monitoringpb.TimeSeriesData, error) {
	ctx := context.Background()

	jwt, err := getJwtConfig(g.keyByte, g.scope)
	if err != nil {
		return nil, err
	}

	c, err := queryClient(ctx, jwt.TokenSource(ctx))
	if err != nil {
		return nil, err
	}

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
			return nil, fmt.Errorf("could not list time series: %w", err)
		}
		result = append(result, resp)
	}

	defer c.Close()

	return result, nil
}

// The function initializes a query client for Google Cloud Monitoring using a token source.
func queryClient(ctx context.Context, ts oauth2.TokenSource) (*monitoring.QueryClient, error) {
	c, err := monitoring.NewQueryClient(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("could not initialize query client <%w>", err)
	}

	return c, nil
}
