package err

import (
	"fmt"
	"net/http"
	"strconv"
)

func CheckHeader(header *http.Header) error {
	xEresultStr := header.Get("X-Eresult")
	if xEresultStr == "" {
		return nil
	}
	xEresult, err := strconv.Atoi(xEresultStr)
	if err != nil {
		return err
	}
	if xEresult != 1 {
		codeMsg, ok := codeMap[xEresult]
		if !ok {
			return fmt.Errorf("fail to login, error: Unkown error code %d", xEresult)
		}
		return fmt.Errorf("fail to login, error: %s", codeMsg)
	}
	return nil
}
