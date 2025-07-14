package export

import (
	"context"
	"fmt"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/whosonfirst/go-whosonfirst-export/v3/properties"
	wof_properties "github.com/whosonfirst/go-whosonfirst-feature/properties"
)

func SupersedeRecord(ctx context.Context, ex Exporter, old_body []byte) ([]byte, []byte, error) {

	id_rsp := gjson.GetBytes(old_body, wof_properties.PATH_WOF_ID)

	if !id_rsp.Exists() {
		return nil, nil, fmt.Errorf("failed to derive old properties.wof:id property for record being superseded")
	}

	old_id := id_rsp.Int()

	// Create the new record

	new_body := old_body

	new_body, err := sjson.DeleteBytes(new_body, wof_properties.PATH_WOF_ID)

	if err != nil {
		return nil, nil, properties.RemovePropertyFailed(wof_properties.PATH_WOF_ID, err)
	}

	_, new_body, err = ex.Export(ctx, new_body)

	if err != nil {
		return nil, nil, err
	}

	id_rsp = gjson.GetBytes(new_body, wof_properties.PATH_WOF_ID)

	if !id_rsp.Exists() {
		return nil, nil, fmt.Errorf("failed to derive new properties.wof:id property for record superseding '%d'", old_id)
	}

	new_id := id_rsp.Int()

	// Update the new record

	new_body, err = sjson.SetBytes(new_body, wof_properties.PATH_WOF_SUPERSEDES, []int64{old_id})

	if err != nil {
		return nil, nil, properties.SetPropertyFailed(wof_properties.PATH_WOF_SUPERSEDES, err)
	}

	// Update the old record

	to_update := map[string]interface{}{
		wof_properties.PATH_MZ_ISCURRENT:      0,
		wof_properties.PATH_WOF_SUPERSEDED_BY: []int64{new_id},
	}

	old_body, err = AssignProperties(ctx, old_body, to_update)

	if err != nil {
		return nil, nil, err
	}

	return old_body, new_body, nil
}

func SupersedeRecordWithParent(ctx context.Context, ex Exporter, to_supersede_f []byte, parent_f []byte) ([]byte, []byte, error) {

	id_rsp := gjson.GetBytes(parent_f, wof_properties.PATH_WOF_ID)

	if !id_rsp.Exists() {
		return nil, nil, fmt.Errorf("parent feature is missing %s", wof_properties.PATH_WOF_ID)
	}

	parent_id := id_rsp.Int()

	hier_rsp := gjson.GetBytes(parent_f, wof_properties.PATH_WOF_HIERARCHY)

	if !hier_rsp.Exists() {
		return nil, nil, fmt.Errorf("parent feature is missing %s", wof_properties.PATH_WOF_HIERARCHY)
	}

	parent_hierarchy := hier_rsp.Value()

	inception_rsp := gjson.GetBytes(parent_f, wof_properties.PATH_EDTF_INCEPTION)

	if !inception_rsp.Exists() {
		return nil, nil, fmt.Errorf("parent record is missing %s", wof_properties.PATH_EDTF_INCEPTION)
	}

	cessation_rsp := gjson.GetBytes(parent_f, wof_properties.PATH_EDTF_CESSATION)

	if !cessation_rsp.Exists() {
		return nil, nil, fmt.Errorf("parent record is missing %s", wof_properties.PATH_EDTF_CESSATION)
	}

	inception := inception_rsp.String()
	cessation := cessation_rsp.String()

	to_update_old := map[string]interface{}{
		wof_properties.PATH_EDTF_INCEPTION: inception,
	}

	to_update_new := map[string]interface{}{
		wof_properties.PATH_WOF_PARENTID:   parent_id,
		wof_properties.PATH_WOF_HIERARCHY:  parent_hierarchy,
		wof_properties.PATH_EDTF_INCEPTION: inception,
		wof_properties.PATH_EDTF_CESSATION: cessation,
	}

	//

	superseded_f, superseding_f, err := SupersedeRecord(ctx, ex, to_supersede_f)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to supersede record: %v", err)
	}

	superseded_f, err = AssignProperties(ctx, superseded_f, to_update_old)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to assign properties for new record: %v", err)
	}

	name_rsp := gjson.GetBytes(superseding_f, wof_properties.PATH_WOF_NAME)

	if !name_rsp.Exists() {
		return nil, nil, fmt.Errorf("failed to retrieve %sfor new record", wof_properties.PATH_WOF_NAME)
	}

	name := name_rsp.String()
	label := fmt.Sprintf("%s (%s)", name, inception)

	to_update_new[wof_properties.PATH_WOF_LABEL] = label

	superseding_f, err = AssignProperties(ctx, superseding_f, to_update_new)

	if err != nil {
		return nil, nil, fmt.Errorf("failed to assign updated properties for new record %v", err)
	}

	return superseded_f, superseding_f, nil

}
