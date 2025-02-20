package liner

import (
	"errors"
	"fmt"
	"syscall"
	"time"
)

const eintrMsg = "interrupted system call"

//
// SetTimeout initializes the timeout value (seconds from 1 to 32767)
// that the Prompt routine will wait for before aborting with a timeout
// error.  A timeout of 0 (default) means no timeout
//

func (s *State) SetTimeout(timeout int16) error {

	var zeroTime time.Time

	if timeout < 0 {
		return ErrInvalidTimeout
	} else if timeout == 0 {
		s.timeout = 0
		s.deadline = zeroTime
		return nil
	} else {
		s.timeout = timeout
		s.deadline = time.Now()
		s.deadline = s.deadline.Add(time.Second * time.Duration(timeout))
		return nil
	}
}

func (s *State) pollStdin() error {

	var err error
	var rdset syscall.FdSet
	var n, secs int

	if s.timeout == 0 {
		return nil
	}

	for {
		now := time.Now()
		if secs = int(s.deadline.Sub(now).Seconds()); secs <= 0 {
			return nil
		}

		to := &syscall.Timeval{Sec: int64(secs), Usec: 0}
		rdset.Bits[syscall.Stdin] = 1
		if n, err = syscall.Select(int(syscall.Stdin)+1, &rdset,
			nil, nil, to); err != nil {
			if err.Error() == eintrMsg {
				continue
			} else {
				return err
			}
		} else if n == 0 {
			return errTimedOut
		} else if n != 1 {
			return errors.New(fmt.Sprintf("n == %d", n))
		}

		if rdset.Bits[syscall.Stdin] == 0 {
			return errors.New(fmt.Sprintf("fd %d not ready", n))
		} else {
			return nil
		}
	}
}
