package detector

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/daniellavrushin/b4/log"
)

const (
	tcpChunkSize = 128 // Read in small chunks to detect exact drop point
	tcpOkThreshold = 70 * 1024  // 70KB — if we get this much, endpoint is OK
)

func (s *DetectorSuite) runTCPCheck(ctx context.Context) *TCPResult {
	log.DiscoveryLogf("[Detector] Starting TCP 16-20KB connection drop test for %d targets", len(TCPTargets))

	result := &TCPResult{}
	results := make([]TCPTargetResult, len(TCPTargets))

	var wg sync.WaitGroup
	sem := make(chan struct{}, 15)

	for i, target := range TCPTargets {
		if s.isCanceled() {
			break
		}
		wg.Add(1)
		go func(idx int, tgt TCPTarget) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			tr := s.testTCPTarget(ctx, tgt)
			results[idx] = tr

			s.mu.Lock()
			s.CompletedChecks++
			s.mu.Unlock()
		}(i, target)
	}

	wg.Wait()

	for _, tr := range results {
		result.Targets = append(result.Targets, tr)
		switch tr.Status {
		case TCPOk:
			result.OkCount++
		case TCPDetected:
			result.DetectedCount++
		}
	}

	total := len(results)
	result.Summary = fmt.Sprintf("%d/%d OK, %d TSPU drops detected",
		result.OkCount, total, result.DetectedCount)

	if result.DetectedCount > 0 {
		pct := float64(result.DetectedCount) / float64(total) * 100
		result.Summary += fmt.Sprintf(" (%.0f%% of endpoints affected)", pct)
	}

	log.DiscoveryLogf("[Detector] TCP check complete: %s", result.Summary)
	return result
}

func (s *DetectorSuite) testTCPTarget(ctx context.Context, target TCPTarget) TCPTargetResult {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			d := net.Dialer{Timeout: 10 * time.Second}
			return d.DialContext(ctx, network, addr)
		},
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", target.URL, nil)
	if err != nil {
		return TCPTargetResult{
			Target: target,
			Status: TCPError,
			Detail: err.Error(),
		}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := client.Do(req)
	if err != nil {
		status, detail := ClassifyTCPError(err, 0)
		return TCPTargetResult{
			Target: target,
			Status: status,
			Detail: detail,
		}
	}
	defer resp.Body.Close()

	// Stream data in small chunks, counting bytes
	var totalRead int64
	buf := make([]byte, tcpChunkSize)

	for {
		n, err := resp.Body.Read(buf)
		totalRead += int64(n)

		if totalRead >= tcpOkThreshold {
			// Successfully read enough data — no TSPU drop
			return TCPTargetResult{
				Target:    target,
				Status:    TCPOk,
				BytesRead: totalRead,
				Detail:    fmt.Sprintf("Read %s KB successfully", formatKB(float64(totalRead)/1024)),
			}
		}

		if err != nil {
			if err == io.EOF {
				// File ended normally before threshold — that's fine (small file)
				return TCPTargetResult{
					Target:    target,
					Status:    TCPOk,
					BytesRead: totalRead,
					Detail:    fmt.Sprintf("Complete: %s KB", formatKB(float64(totalRead)/1024)),
				}
			}

			status, detail := ClassifyTCPError(err, totalRead)
			return TCPTargetResult{
				Target:    target,
				Status:    status,
				BytesRead: totalRead,
				DropAtKB:  float64(totalRead) / 1024,
				Detail:    detail,
			}
		}

		if n == 0 {
			break
		}
	}

	return TCPTargetResult{
		Target:    target,
		Status:    TCPOk,
		BytesRead: totalRead,
		Detail:    fmt.Sprintf("Read %s KB", formatKB(float64(totalRead)/1024)),
	}
}
