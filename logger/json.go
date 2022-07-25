package lggr

import "encoding/json"

func (log *LogWrapper) UnmarshalJSONSpecial(bytes []byte) error {
	var nl LogWrapper

	err := json.Unmarshal(bytes, &nl)
	if err != nil {
		return err
	}

	*log = *logger.With(nl.DetachedFields...).WithCallerSkip(nl.CallerSkip)

	return nil
}
