package dockertarget

// NOTE: This code is adapted from Promtail (90a1d4593e2d690b37333386383870865fe177bf).
// The dockertarget package is used to configure and run the targets that can
// read logs from Docker containers and forward them to other loki components.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/grafana/agent/component/common/loki/client/fake"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/go-kit/log"
	"github.com/grafana/agent/component/common/loki/positions"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/relabel"
	"github.com/stretchr/testify/require"
)

func TestDockerTarget(t *testing.T) {
	h := func(w http.ResponseWriter, r *http.Request) {
		switch path := r.URL.Path; {
		case strings.HasSuffix(path, "/logs"):
			var filePath string
			if strings.Contains(r.URL.RawQuery, "since=0") {
				filePath = "testdata/flog.log"
			} else {
				filePath = "testdata/flog_after_restart.log"
			}
			dat, err := os.ReadFile(filePath)
			require.NoError(t, err)
			_, err = w.Write(dat)
			require.NoError(t, err)
		default:
			w.Header().Set("Content-Type", "application/json")
			info := types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{},
				Mounts:            []types.MountPoint{},
				Config:            &container.Config{Tty: false},
				NetworkSettings:   &types.NetworkSettings{},
			}
			err := json.NewEncoder(w).Encode(info)
			require.NoError(t, err)
		}
	}

	ts := httptest.NewServer(http.HandlerFunc(h))
	defer ts.Close()

	w := log.NewSyncWriter(os.Stderr)
	logger := log.NewLogfmtLogger(w)
	entryHandler := fake.NewClient(func() {})
	client, err := client.NewClientWithOpts(client.WithHost(ts.URL))
	require.NoError(t, err)

	ps, err := positions.New(logger, positions.Config{
		SyncPeriod:    10 * time.Second,
		PositionsFile: t.TempDir() + "/positions.yml",
	})
	require.NoError(t, err)

	tgt, err := NewTarget(
		NewMetrics(prometheus.NewRegistry()),
		logger,
		entryHandler,
		ps,
		"flog",
		model.LabelSet{"job": "docker"},
		[]*relabel.Config{},
		client,
	)
	require.NoError(t, err)
	tgt.StartIfNotRunning()

	require.Eventually(t, func() bool {
		return len(entryHandler.Received()) >= 5
	}, 5*time.Second, 100*time.Millisecond)

	received := entryHandler.Received()
	sort.Slice(received, func(i, j int) bool {
		return received[i].Timestamp.Before(received[j].Timestamp)
	})

	expectedLines := []string{
		"5.3.69.55 - - [09/Dec/2021:09:15:02 +0000] \"HEAD /brand/users/clicks-and-mortar/front-end HTTP/2.0\" 503 27087",
		"101.54.183.185 - - [09/Dec/2021:09:15:03 +0000] \"POST /next-generation HTTP/1.0\" 416 11468",
		"69.27.137.160 - runolfsdottir2670 [09/Dec/2021:09:15:03 +0000] \"HEAD /content/visionary/engineer/cultivate HTTP/1.1\" 302 2975",
		"28.104.242.74 - - [09/Dec/2021:09:15:03 +0000] \"PATCH /value-added/cultivate/systems HTTP/2.0\" 405 11843",
		"150.187.51.54 - satterfield1852 [09/Dec/2021:09:15:03 +0000] \"GET /incentivize/deliver/innovative/cross-platform HTTP/1.1\" 301 13032",
	}
	actualLines := make([]string, 0, 5)
	for _, entry := range received[:5] {
		actualLines = append(actualLines, entry.Line)
	}
	require.ElementsMatch(t, actualLines, expectedLines)

	// restart target to simulate container restart
	tgt.StartIfNotRunning()
	entryHandler.Clear()
	require.Eventually(t, func() bool {
		return len(entryHandler.Received()) >= 5
	}, 5*time.Second, 100*time.Millisecond)

	receivedAfterRestart := entryHandler.Received()
	sort.Slice(receivedAfterRestart, func(i, j int) bool {
		return receivedAfterRestart[i].Timestamp.Before(receivedAfterRestart[j].Timestamp)
	})
	actualLinesAfterRestart := make([]string, 0, 5)
	for _, entry := range receivedAfterRestart[:5] {
		actualLinesAfterRestart = append(actualLinesAfterRestart, entry.Line)
	}
	expectedLinesAfterRestart := []string{
		"243.115.12.215 - - [09/Dec/2023:09:16:57 +0000] \"DELETE /morph/exploit/granular HTTP/1.0\" 500 26468",
		"221.41.123.237 - - [09/Dec/2023:09:16:57 +0000] \"DELETE /user-centric/whiteboard HTTP/2.0\" 205 22487",
		"89.111.144.144 - - [09/Dec/2023:09:16:57 +0000] \"DELETE /open-source/e-commerce HTTP/1.0\" 401 11092",
		"62.180.191.187 - - [09/Dec/2023:09:16:57 +0000] \"DELETE /cultivate/integrate/technologies HTTP/2.0\" 302 12979",
		"156.249.2.192 - - [09/Dec/2023:09:16:57 +0000] \"POST /revolutionize/mesh/metrics HTTP/2.0\" 401 5297",
	}
	require.ElementsMatch(t, actualLinesAfterRestart, expectedLinesAfterRestart)
}
