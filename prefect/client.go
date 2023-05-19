package prefect

import (
	"fmt"
	hc "github.com/prefecthq/terraform-provider-prefect/prefect_api"
)

func getClient(m interface{}) (*hc.Client, error) {
	c, ok := m.(*hc.Client)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T for m", m)
	}
	return c, nil
}
