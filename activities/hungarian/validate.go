package hungarian

// Adapted from https://github.com/oddg/hungarian-algorithm

import (
	"errors"
	"fmt"
)

func validate(costs [][]float64) error {
	n := len(costs)

	if n == 0 {
		return errors.New("The costs matrix is empty.")
	}

	if m := len(costs[0]); m != n {
		return errors.New("The costs matrix is not square.")
	}

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if costs[i][j] < 0 {
				return fmt.Errorf("The coefficient (%d,%d) is negative.", i, j)
			}
		}
	}

	return nil
}
