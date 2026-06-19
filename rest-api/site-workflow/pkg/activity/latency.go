// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package activity

import (
	"time"

	"github.com/rs/zerolog"
)

func logGrpcCallLatency(logger *zerolog.Logger, operation string, duration time.Duration, err error) {
	if err != nil {
		logger.Warn().Err(err).Dur("duration", duration).Msgf("Failed to %s using Site Controller API", operation)
	} else {
		logger.Debug().Dur("duration", duration).Msgf("Completed %s", operation)
	}
}
