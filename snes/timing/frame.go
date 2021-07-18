package timing

import "time"

// Frame = 5,369,317.5 / 89,341.5 ~= 60.0988062658451 frames / sec ~= 16,639,265.605 ns / frame
const Frame = time.Nanosecond * 16_639_265
