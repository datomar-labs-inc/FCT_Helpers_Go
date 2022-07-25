package lggr

import "encoding/json"

func (log *LogWrapper) MarshalJSON() ([]byte, error) {
	return json.Marshal(log)
}

func (log *LogWrapper) UnmarshalJSON(bytes []byte) error {
	var nl LogWrapper

	err := json.Unmarshal(bytes, &nl)
	if err != nil {
		return err
	}

	*log = *logger.With(nl.DetachedFields...).WithCallerSkip(nl.CallerSkip)

	return nil
}
