// Package stunx contains [github.com/pion/stun] extensions.
package stunx

import (
	"errors"
	"net"
	"os"
	"time"

	"github.com/ooni/probe-engine/pkg/model"
	"github.com/ooni/probe-engine/pkg/runtimex"
	"github.com/pion/stun"
)

// transactionConfig contains configuration for the STUN transaction.
type transactionConfig struct {
	// logger is the logger to use.
	logger model.Logger

	// rc is the maximum number of times we're allowed to send.
	rc int

	// rm is the number of rtos to wait before declaring the transaction failed.
	rm int

	// rto is the retransmission timeout to use.
	rto time.Duration
}

// TransactionOption allows to configure the STUN transaction algorithm.
type TransactionOption func(tc *transactionConfig)

// TransactionOptionLogger sets the logger.
func TransactionOptionLogger(logger model.Logger) TransactionOption {
	return func(tc *transactionConfig) {
		tc.logger = logger
	}
}

// TransactionOptionRm sets the Rm parameter (i.e., the number or RTOs to wait after the last send).
//
// Note that this function PANICS if passed a zero or negative value.
func TransactionOptionRm(rm int) TransactionOption {
	runtimex.Assert(rm > 0, "TransactionOptionRm passed a zero or negative value")
	return func(tc *transactionConfig) {
		tc.rm = rm
	}
}

// TransactionOptionRTO sets the RTO for the transaction.
//
// Note that this function PANICS if passed a zero or negative value.
func TransactionOptionRTO(rto time.Duration) TransactionOption {
	runtimex.Assert(rto > 0, "TransactionOptionRTO passed a zero or negative value")
	return func(tc *transactionConfig) {
		tc.rto = rto
	}
}

// TransactionOptionRc sets the Rc parameter (i.e., the maximum number of transmissions to attempt).
//
// Note that this function PANICS if passed a zero or negative value.
func TransactionOptionRc(rc int) TransactionOption {
	runtimex.Assert(rc > 0, "TransactionOptionRc passed a zero or negative value")
	return func(tc *transactionConfig) {
		tc.rc = rc
	}
}

// RunBindingRequestTransaction runs a binding-request STUN transaction using the given  conn, and options.
func RunBindingRequestTransaction(conn net.Conn, options ...TransactionOption) (*stun.Message, error) {
	// set https://datatracker.ietf.org/doc/html/rfc5389#section-7.2.1 defaults
	config := &transactionConfig{
		logger: model.DiscardLogger,
		rc:     7,
		rm:     16,
		rto:    500 * time.Millisecond,
	}

	// make a binding request
	req := stun.MustBuild(stun.TransactionID, stun.BindingRequest)

	// honour options
	for _, option := range options {
		option(config)
	}

	// sender loop
	for idx := 0; idx < config.rc; idx++ {
		// typically the timeout equals the RTO
		timeout := config.rto

		// however the final timeout is Rm times the RTO
		if idx == config.rc-1 {
			timeout = time.Duration(config.rm) * config.rto
		}

		// we can now double the RTO
		config.rto *= 2

		// configure the deadline accordingly
		if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
			config.logger.Debugf("stunx: cannot set conn deadline: %s", err.Error())
			return nil, err
		}

		// send the request to the remote host
		if _, err := req.WriteTo(conn); err != nil {
			config.logger.Debugf("stunx: conn.Write failed: %s", err.Error())
			return nil, err
		}

		for {
			// receive the response or time out
			const respBufferSize = 4 << 10
			resp := &stun.Message{}
			resp.Raw = make([]byte, respBufferSize)
			_, err := resp.ReadFrom(conn)

			// handle the timeout case by jumping to the next iteration
			if errors.Is(err, os.ErrDeadlineExceeded) {
				break
			}

			// handle the case of a hard error
			if err != nil {
				config.logger.Debugf("stunx: conn.Read failed: %s", err.Error())
				return nil, err
			}

			// process message according to https://datatracker.ietf.org/doc/html/rfc5389#section-7.3
			// thus ignoring messages with unexpected type or transaction ID.
			if resp.Type != stun.BindingSuccess && resp.Type != stun.BindingError {
				continue
			}
			if resp.TransactionID != req.TransactionID {
				continue
			}

			// pass good response to the caller
			return resp, nil
		}
	}

	// surrender and declare timeout
	return nil, stun.ErrTransactionTimeOut
}
