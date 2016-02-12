package lock

import (
	"time"

	redis "gopkg.in/redis.v3"
)

var (
	redisClient *redis.Client
)

func Init(client *redis.Client) {
	redisClient = client
}

//since Redis doesn't support blocking SET operation
//thus we need to repeatingly probe the redis to perform SETNX operaion
//this function return the sleepTime between each operation
func waitTimeSeries(waitTime time.Duration) []time.Duration {
	beginingMs := 800 * time.Millisecond
	endingMs := 50 * time.Millisecond

	temp := []time.Duration{0}

	t := endingMs
	total := 0 * time.Millisecond
	for t < beginingMs && total+t <= waitTime {
		temp = append(temp, t)
		total = total + t
		t = time.Duration(float64(t) * 1.2)
	}

	for total+beginingMs < waitTime {
		temp = append(temp, beginingMs)
		total = total + beginingMs
	}

	if t := waitTime - total; t > 0 {
		temp = append(temp, t)
	}

	//reverse the series
	output := []time.Duration{}
	for i, _ := range temp {
		output = append(output, temp[len(temp)-i-1])
	}

	return output
}

//lockDuration = the total effective period for the lock
//waitTime = if the lock is already hold by someone, the period of time that current thread should wait for
func AcquireLock(name string, lockDuration, waitTime time.Duration) (bool, error) {
	for _, t := range waitTimeSeries(waitTime) {
		ok, err := redisClient.SetNX(name, ``, lockDuration).Result()
		if err != nil {
			return false, err
		}

		if ok {
			return true, nil
		} else {
			if t >= 0 {
				time.Sleep(t)
			} else {
				break
			}
		}
	}
	return false, nil
}

func ReleaseLock(name string) error {
	_, err := redisClient.Del(name).Result()
	return err
}
