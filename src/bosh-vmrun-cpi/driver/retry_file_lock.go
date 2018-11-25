package driver

import (
	"errors"
	"time"

	boshlog "github.com/cloudfoundry/bosh-utils/logger"
	"github.com/gofrs/flock"
)

type retryFileLockImpl struct {
	logger boshlog.Logger
}

const (
	pollInterval = 250 * time.Millisecond
)

func NewRetryFileLock(logger boshlog.Logger) RetryFileLock {
	return &retryFileLockImpl{logger: logger}
}

func (c retryFileLockImpl) Try(lockFilePath string, maxWait time.Duration, fn func() error) error {
	fileLock := flock.New(lockFilePath)

	for i := time.Duration(0); i < maxWait; i += pollInterval {
		locked, err := fileLock.TryLock()
		if err != nil {
			c.logger.Error("retry-file-lock", "lock failed for %s", lockFilePath)
			return err
		}
		defer fileLock.Unlock()

		if locked {
			c.logger.Debug("retry-file-lock", "lock acquired: %s", lockFilePath)

			return fn()
		} else {
			c.logger.Debug("retry-file-lock", "lock not acquired, waiting: %s", lockFilePath)
			time.Sleep(pollInterval)
			continue
		}
	}

	return errors.New("failed to exit")
}
