package lggr

import "encoding/json"

func (log *LogWrapper) UnmarshalJSONSpecial(bytes []byte) error {
	var nl LogWrapper

	err := json.Unmarshal(bytes, &nl)
	if err != nil {
		return err
	}

	*log = *log.With(nl.DetachedFields...).WithCallerSkip(nl.CallerSkip)

	return nil
}
