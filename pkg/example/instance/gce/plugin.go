package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/infrakit/pkg/example/instance/gce/gcloud"
	"github.com/docker/infrakit/pkg/spi/instance"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

type instanceProperties struct {
	NamePrefix  string
	Description string
	MachineType string
	Network     string
	DiskSizeMb  int64
	Tags        []string
}

type gceInstance struct {
	instance.Description
}

type plugin struct {
	api gcloud.GCloud
}

// NewGCEInstancePlugin creates a new GCE instance plugin for a given project
// and zone.
func NewGCEInstancePlugin(project, zone string) instance.Plugin {
	log.Debugln("gce instance plugin. project=", project)

	api, err := gcloud.New(project, zone)
	if err != nil {
		panic(err)
	}

	return &plugin{
		api: api,
	}
}

func parseProperties(properties json.RawMessage) (*instanceProperties, error) {
	p := instanceProperties{}

	if err := json.Unmarshal(properties, &p); err != nil {
		return nil, err
	}

	if p.NamePrefix == "" {
		p.NamePrefix = "instance"
	}
	if p.DiskSizeMb == 0 {
		p.DiskSizeMb = 10
	}

	return &p, nil
}

func (p *plugin) Validate(req json.RawMessage) error {
	log.Debugln("validate", string(req))

	instanceProperties, err := parseProperties(req)
	if err != nil {
		return err
	}

	missingProperties := []string{}
	if instanceProperties.MachineType == "" {
		missingProperties = append(missingProperties, "MachineType")
	}
	if instanceProperties.Network == "" {
		missingProperties = append(missingProperties, "Network")
	}

	switch len(missingProperties) {
	case 0:
		return nil
	default:
		return fmt.Errorf("Missing: %s", strings.Join(missingProperties, ", "))
	}
}

func (p *plugin) Provision(spec instance.Spec) (*instance.ID, error) {
	properties, err := parseProperties(*spec.Properties)
	if err != nil {
		return nil, err
	}

	name := fmt.Sprintf("%s-%d", properties.NamePrefix, rand.Int63())
	id := instance.ID(name)

	err = p.api.CreateInstance(name, &gcloud.InstanceSettings{
		Description: properties.Description,
		MachineType: properties.MachineType,
		Network:     properties.Network,
		Tags:        properties.Tags,
		DiskSizeMb:  properties.DiskSizeMb,
		MetaData:    gcloud.TagsToMetaData(spec.Tags),
	})

	log.Debugln("provision", id, "err=", err)
	if err != nil {
		return nil, err
	}

	return &id, nil
}

func (p *plugin) Destroy(id instance.ID) error {
	err := p.api.DeleteInstance(string(id))
	log.Debugln("destroy", id, "err=", err)

	return err
}

func (p *plugin) DescribeInstances(tags map[string]string) ([]instance.Description, error) {
	log.Debugln("describe-instances", tags)

	instances, err := p.api.ListInstances()
	if err != nil {
		return nil, err
	}

	log.Debugln("total count:", len(instances))

	result := []instance.Description{}

scan:
	for _, inst := range instances {
		instTags := gcloud.MetaDataToTags(inst.Metadata.Items)

		for k, v := range tags {
			if instTags[k] != v {
				continue scan // we implement AND
			}
		}

		result = append(result, instance.Description{
			ID:   instance.ID(inst.Name),
			Tags: instTags,
		})
	}

	log.Debugln("matching count:", len(result))

	return result, nil
}
