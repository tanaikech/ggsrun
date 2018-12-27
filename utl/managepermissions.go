// Package utl (managepermissions.go) :
// These methods are for managing permissions of files and folders in Google Drive.
package utl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
)

// permissionInf : Struct for permission information.
type permissionInf struct {
	FileID            string `json:"fileId,omitempty"`
	PermissionID      string `json:"permissionId,omitempty"`
	Role              string `json:"role,omitempty"`
	Type              string `json:"type,omitempty"`
	Emailaddress      string `json:"emailaddress,omitempty"`
	Transferownership bool   `json:"transferownership,omitempty"`
	CreateObject      string `json:"createRequestBody,omitempty"`
	DeleteObject      string `json:"deleteRequestBody,omitempty"`
	UpdateObject      string `json:"updateRequestBody,omitempty"`
	Create            bool   `json:"-"`
	Delete            bool   `json:"-"`
}

// getPermissions : Retrieve permissions.
func (p *FileInf) getPermissions(u *url.URL) error {
	u.Path = path.Join(u.Path, p.PermissionInfo.PermissionID)
	q := u.Query()
	q.Set("fields", "allowFileDiscovery,deleted,displayName,domain,emailAddress,expirationTime,id,kind,photoLink,role,teamDrivePermissionDetails,type")
	u.RawQuery = q.Encode()
	r := &RequestParams{
		Method:      "GET",
		APIURL:      u.String(),
		Data:        nil,
		Accesstoken: p.Accesstoken,
		Dtime:       30,
	}
	p.reqAndGetRawResponse(r)
	return nil
}

// getPermissionsList : Retrieve permissions list.
func (p *FileInf) getPermissionsList(u *url.URL) error {
	q := u.Query()
	q.Set("pageSize", "100")
	q.Set("fields", "kind,nextPageToken,permissions")
	u.RawQuery = q.Encode()
	r := &RequestParams{
		Method:      "GET",
		APIURL:      u.String(),
		Data:        nil,
		Accesstoken: p.Accesstoken,
		Dtime:       30,
	}
	p.reqAndGetRawResponse(r)
	return nil
}

// deletePermissions : Delete permissions.
func (p *FileInf) deletePermissions(u *url.URL) error {
	u.Path = path.Join(u.Path, p.PermissionInfo.PermissionID)
	r := &RequestParams{
		Method:      "DELETE",
		APIURL:      u.String(),
		Accesstoken: p.Accesstoken,
		Dtime:       30,
	}
	_, err := r.FetchAPIRaw()
	if err != nil {
		return fmt.Errorf("Permission of '%s' was NOT deleted. Please confirm the fileId and permissionId again", p.PermissionInfo.PermissionID)
	}
	p.Msgar = append(p.Msgar, fmt.Sprintf("Permission of '%s' was deleted.", p.PermissionInfo.PermissionID))
	return nil
}

// createPermissions : Create permissions.
func (p *FileInf) createPermissions(u *url.URL) error {
	q := u.Query()
	if p.PermissionInfo.Role == "" || p.PermissionInfo.Type == "" {
		return fmt.Errorf("role and type are required for creating permissions")
	}
	if p.PermissionInfo.Transferownership {
		q.Set("transferOwnership", "true")
	}
	q.Set("fields", "allowFileDiscovery,deleted,displayName,domain,emailAddress,expirationTime,id,kind,photoLink,role,teamDrivePermissionDetails,type")
	u.RawQuery = q.Encode()
	meta := struct {
		Role         string `json:"role,omitempty"`
		Type         string `json:"type,omitempty"`
		EmailAddress string `json:"emailAddress,omitempty"`
	}{
		p.PermissionInfo.Role, p.PermissionInfo.Type, p.PermissionInfo.Emailaddress,
	}
	metadata, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	r := &RequestParams{
		Method:      "POST",
		APIURL:      u.String(),
		Data:        bytes.NewBuffer(metadata),
		Accesstoken: p.Accesstoken,
		Contenttype: "application/json",
		Dtime:       30,
	}
	p.reqAndGetRawResponse(r)
	p.Msgar = append(p.Msgar, "Permission was created.")
	return nil
}

// getURL : Get URL for fetching Permissions.
func (p *FileInf) getURL() (*url.URL, error) {
	u, err := url.Parse(driveapiurl)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, p.PermissionInfo.FileID)
	u.Path = path.Join(u.Path, "permissions")
	return u, nil
}

// ManagePermissions : Main method of Manage Permissions.
func (p *FileInf) ManagePermissions() *FileInf {
	var err error
	u, err := p.getURL()
	if p.PermissionInfo.FileID != "" {
		if p.PermissionInfo.PermissionID == "" {
			if (p.PermissionInfo.Create && !p.PermissionInfo.Delete) || (p.PermissionInfo.CreateObject != "" && p.PermissionInfo.DeleteObject == "" && p.PermissionInfo.UpdateObject == "") {
				err = p.createPermissions(u)
			} else {
				err = p.getPermissionsList(u)
			}
		} else {
			if (!p.PermissionInfo.Create && p.PermissionInfo.Delete) || (p.PermissionInfo.CreateObject == "" && p.PermissionInfo.DeleteObject != "" && p.PermissionInfo.UpdateObject == "") {
				err = p.deletePermissions(u)
			} else {
				err = p.getPermissions(u)
			}
		}
	} else {
		err = fmt.Errorf("Invalid options. Please check HELP using $ ggsrun p --help")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return p
}
