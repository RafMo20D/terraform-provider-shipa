package client

import "strconv"

type UpdateAppRequest struct {
	Pool        string   `json:"pool,omitempty"`
	TeamOwner   string   `json:"teamowner,omitempty"`
	Description string   `json:"description,omitempty"`
	Plan        string   `json:"plan,omitempty"`
	Platform    string   `json:"platform,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

func NewUpdateAppRequest(a *App) *UpdateAppRequest {
	var plan string
	if a.Plan != nil {
		plan = a.Plan.Name
	}

	return &UpdateAppRequest{
		Pool:        a.Pool,
		TeamOwner:   a.TeamOwner,
		Description: a.Description,
		Plan:        plan,
		Platform:    a.Platform,
		Tags:        a.Tags,
	}
}

type App struct {
	Name        string        `json:"name,omitempty"`
	Description string        `json:"description,omitempty"`
	Pool        string        `json:"pool" terraform:"framework"`
	TeamOwner   string        `json:"teamowner,omitempty"`
	Plan        *Plan         `json:"plan,omitempty"`
	Units       []*Unit       `json:"units,omitempty"`
	Cname       []string      `json:"cname,omitempty"`
	IP          string        `json:"ip,omitempty"`
	Org         string        `json:"org,omitempty"`
	Entrypoints []*Entrypoint `json:"entrypoints,omitempty"`
	Routers     []*Router     `json:"routers,omitempty"`
	Lock        *Lock         `json:"lock,omitempty"`
	Tags        []string      `json:"tags,omitempty"`
	Platform    string        `json:"platform,omitempty"`
	Status      string        `json:"status,omitempty"`
	Error       string        `json:"error,omitempty"` // not shows in API response
}

type Plan struct {
	Name     string   `json:"name,omitempty"`
	Memory   int64    `json:"memory"`
	Swap     int64    `json:"swap"`
	CPUShare int64    `json:"cpushare"`
	Default  bool     `json:"default"`
	Public   bool     `json:"public"`
	Org      string   `json:"org,omitempty"`
	Teams    []string `json:"teams,omitempty"`
}

type Unit struct {
	ID          string   `json:"ID,omitempty"`
	Name        string   `json:"Name,omitempty"`
	AppName     string   `json:"AppName,omitempty"`
	ProcessName string   `json:"ProcessName,omitempty"`
	Type        string   `json:"Type,omitempty"`
	IP          string   `json:"IP,omitempty"`
	Status      string   `json:"Status,omitempty"`
	Version     string   `json:"Version,omitempty"`
	Org         string   `json:"Org,omitempty"`
	HostAddr    string   `json:"HostAddr,omitempty"`
	HostPort    string   `json:"HostPort,omitempty"`
	Address     *Address `json:"Address,omitempty"`
}

type Address struct {
	Scheme      string `json:"Scheme,omitempty"`
	Host        string `json:"Host,omitempty"`
	Opaque      string `json:"Opaque,omitempty"`
	User        string `json:"User,omitempty"`
	Path        string `json:"Path,omitempty"`
	RawPath     string `json:"RawPath,omitempty"`
	ForceQuery  bool   `json:"ForceQuery"`
	RawQuery    string `json:"RawQuery,omitempty"`
	Fragment    string `json:"Fragment,omitempty"`
	RawFragment string `json:"RawFragment,omitempty"`
}

type Entrypoint struct {
	Cname  string `json:"cname,omitempty"`
	Scheme string `json:"scheme,omitempty"`
}

type Router struct {
	Name    string                 `json:"name,omitempty"`
	Opts    map[string]interface{} `json:"opts,omitempty"`
	Type    string                 `json:"type,omitempty"`
	Address string                 `json:"address,omitempty"`
	Default bool                   `json:"default"` // not show in API response
}

type Lock struct {
	Locked      bool   `json:"Locked"`
	Reason      string `json:"Reason,omitempty"`
	Owner       string `json:"Owner,omitempty"`
	AcquireDate string `json:"AcquireDate,omitempty"`
}

func (c *Client) ListApps() ([]*App, error) {
	apps := make([]*App, 0)
	err := c.get(&apps, apiApps)
	if err != nil {
		return nil, err
	}

	return apps, nil
}

func (c *Client) GetApp(name string) (*App, error) {
	app := &App{}
	err := c.get(app, apiApps, name)
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (c *Client) CreateApp(app *App) error {
	return c.post(app, apiApps)
}

func (c *Client) UpdateApp(name string, app *UpdateAppRequest) error {
	return c.put(app, apiApps, name)
}

func (c *Client) DeleteApp(name string) error {
	return c.delete(apiApps, name)
}

type AppEnv struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type CreateAppEnv struct {
	Envs      []*AppEnv `json:"envs"`
	NoRestart bool      `json:"norestart"`
	Private   bool      `json:"private"`
}

func (c *Client) CreateAppEnvs(appName string, req *CreateAppEnv) error {
	return c.post(req, apiAppEnvs(appName))
}

func (c *Client) GetAppEnvs(appName string) ([]*AppEnv, error) {
	envs := make([]*AppEnv, 0)
	err := c.get(&envs, apiAppEnvs(appName))
	if err != nil {
		return nil, err
	}

	return envs, nil
}

func (c *Client) DeleteAppEnvs(appName string, req *CreateAppEnv) error {
	params := []*QueryParam{
		{Key: "norestart", Val: req.NoRestart},
	}
	for _, p := range req.Envs {
		params = append(params, &QueryParam{Key: "env", Val: p.Name})
	}

	if len(params) > 1 {
		return c.deleteWithParams(params, apiAppEnvs(appName))
	}

	return nil
}

type AppCname struct {
	Cname   string `json:"cname"`
	Encrypt bool   `json:"encrypt"`
}

func (c *Client) CreateAppCname(appName string, req *AppCname) error {
	return c.post(req, apiAppCname(appName))
}

func (c *Client) UpdateAppCname(appName string, req *AppCname) error {
	return c.put(req, apiAppCname(appName))
}

func (c *Client) DeleteAppCname(appName string, req *AppCname) error {
	return c.deleteWithPayload(req, nil, apiAppCname(appName))
}

type AppDeploy struct {
	Image          string `json:"image"`
	PrivateImage   bool   `json:"private-image,omitempty" terraform:"private_image"`
	RegistryUser   string `json:"registry-user,omitempty" terraform:"registry_user"`
	RegistrySecret string `json:"registry-secret,omitempty" terraform:"registry_secret"`
	Steps          int64  `json:"steps,omitempty"`
	StepWeight     int64  `json:"step-weight,omitempty" terraform:"step_weight"`
	StepInterval   int64  `json:"step-interval,omitempty" terraform:"step_interval"`
	Port           int64  `json:"port,omitempty"`
	Detach         bool   `json:"detach"`
	Message        string `json:"message,omitempty"`
}

func (c *Client) DeployApp(appName string, req *AppDeploy) error {
	params := map[string]string{
		"image": req.Image,
	}
	if req.PrivateImage {
		params["private-image"] = "true"
		params["registry-user"] = req.RegistryUser
		params["registry-secret"] = req.RegistrySecret
	}
	if req.Steps > 0 {
		params["steps"] = strconv.FormatInt(req.Steps, 10)
	}
	if req.StepWeight > 0 {
		params["step-weight"] = strconv.FormatInt(req.StepWeight, 10)
	}
	if req.StepInterval > 0 {
		params["step-interval"] = strconv.FormatInt(req.StepInterval, 10)
	}
	if req.Port > 0 {
		params["port"] = strconv.FormatInt(req.Port, 10)
	}
	if req.Detach {
		params["detach"] = "true"
	}
	if req.Message != "" {
		params["message"] = req.Message
	}

	return c.postURLEncoded(params, apiAppDeploy(appName))
}
