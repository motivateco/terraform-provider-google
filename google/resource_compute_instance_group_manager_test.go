package google

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"

	computeBeta "google.golang.org/api/compute/v0.beta"
	"google.golang.org/api/compute/v1"

	"sort"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccInstanceGroupManager_basic(t *testing.T) {
	t.Parallel()

	var manager compute.InstanceGroupManager

	template := fmt.Sprintf("igm-test-%s", acctest.RandString(10))
	target := fmt.Sprintf("igm-test-%s", acctest.RandString(10))
	igm1 := fmt.Sprintf("igm-test-%s", acctest.RandString(10))
	igm2 := fmt.Sprintf("igm-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceGroupManagerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceGroupManager_basic(template, target, igm1, igm2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupManagerExists(
						"google_compute_instance_group_manager.igm-basic", &manager),
					testAccCheckInstanceGroupManagerExists(
						"google_compute_instance_group_manager.igm-no-tp", &manager),
				),
			},
			resource.TestStep{
				ResourceName:      "google_compute_instance_group_manager.igm-basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
			resource.TestStep{
				ResourceName:      "google_compute_instance_group_manager.igm-no-tp",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccInstanceGroupManager_targetSizeZero(t *testing.T) {
	t.Parallel()

	var manager compute.InstanceGroupManager

	templateName := fmt.Sprintf("igm-test-%s", acctest.RandString(10))
	igmName := fmt.Sprintf("igm-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceGroupManagerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceGroupManager_targetSizeZero(templateName, igmName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupManagerExists(
						"google_compute_instance_group_manager.igm-basic", &manager),
				),
			},
		},
	})

	if manager.TargetSize != 0 {
		t.Errorf("Expected target_size to be 0, got %d", manager.TargetSize)
	}
}

func TestAccInstanceGroupManager_update(t *testing.T) {
	t.Parallel()

	var manager compute.InstanceGroupManager

	template1 := fmt.Sprintf("igm-test-%s", acctest.RandString(10))
	target1 := fmt.Sprintf("igm-test-%s", acctest.RandString(10))
	target2 := fmt.Sprintf("igm-test-%s", acctest.RandString(10))
	template2 := fmt.Sprintf("igm-test-%s", acctest.RandString(10))
	igm := fmt.Sprintf("igm-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceGroupManagerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceGroupManager_update(template1, target1, igm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupManagerExists(
						"google_compute_instance_group_manager.igm-update", &manager),
					testAccCheckInstanceGroupManagerUpdated("google_compute_instance_group_manager.igm-update", 2, []string{target1}, template1),
					testAccCheckInstanceGroupManagerNamedPorts(
						"google_compute_instance_group_manager.igm-update",
						map[string]int64{"customhttp": 8080},
						&manager),
				),
			},
			resource.TestStep{
				Config: testAccInstanceGroupManager_update2(template1, target1, target2, template2, igm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupManagerExists(
						"google_compute_instance_group_manager.igm-update", &manager),
					testAccCheckInstanceGroupManagerUpdated(
						"google_compute_instance_group_manager.igm-update", 3,
						[]string{target1, target2}, template2),
					testAccCheckInstanceGroupManagerNamedPorts(
						"google_compute_instance_group_manager.igm-update",
						map[string]int64{"customhttp": 8080, "customhttps": 8443},
						&manager),
				),
			},
		},
	})
}

func TestAccInstanceGroupManager_updateLifecycle(t *testing.T) {
	t.Parallel()

	var manager compute.InstanceGroupManager

	tag1 := "tag1"
	tag2 := "tag2"
	igm := fmt.Sprintf("igm-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceGroupManagerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceGroupManager_updateLifecycle(tag1, igm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupManagerExists(
						"google_compute_instance_group_manager.igm-update", &manager),
				),
			},
			resource.TestStep{
				Config: testAccInstanceGroupManager_updateLifecycle(tag2, igm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupManagerExists(
						"google_compute_instance_group_manager.igm-update", &manager),
					testAccCheckInstanceGroupManagerTemplateTags(
						"google_compute_instance_group_manager.igm-update", []string{tag2}),
				),
			},
		},
	})
}

func TestAccInstanceGroupManager_updateStrategy(t *testing.T) {
	t.Parallel()

	var manager compute.InstanceGroupManager
	igm := fmt.Sprintf("igm-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceGroupManagerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceGroupManager_updateStrategy(igm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupManagerExists(
						"google_compute_instance_group_manager.igm-update-strategy", &manager),
					testAccCheckInstanceGroupManagerUpdateStrategy(
						"google_compute_instance_group_manager.igm-update-strategy", "NONE"),
				),
			},
		},
	})
}

func TestAccInstanceGroupManager_rollingUpdatePolicy(t *testing.T) {
	t.Parallel()

	var manager computeBeta.InstanceGroupManager

	igm := fmt.Sprintf("igm-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceGroupManagerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceGroupManager_rollingUpdatePolicy(igm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupManagerBetaExists(
						"google_compute_instance_group_manager.igm-rolling-update-policy", &manager),
					resource.TestCheckResourceAttr(
						"google_compute_instance_group_manager.igm-rolling-update-policy", "update_strategy", "ROLLING_UPDATE"),
					resource.TestCheckResourceAttr(
						"google_compute_instance_group_manager.igm-rolling-update-policy", "rolling_update_policy.0.type", "PROACTIVE"),
					resource.TestCheckResourceAttr(
						"google_compute_instance_group_manager.igm-rolling-update-policy", "rolling_update_policy.0.minimal_action", "REPLACE"),
					resource.TestCheckResourceAttr(
						"google_compute_instance_group_manager.igm-rolling-update-policy", "rolling_update_policy.0.max_surge_percent", "50"),
					resource.TestCheckResourceAttr(
						"google_compute_instance_group_manager.igm-rolling-update-policy", "rolling_update_policy.0.max_unavailable_percent", "50"),
					resource.TestCheckResourceAttr(
						"google_compute_instance_group_manager.igm-rolling-update-policy", "rolling_update_policy.0.min_ready_sec", "20"),
				),
			},
			resource.TestStep{
				Config: testAccInstanceGroupManager_rollingUpdatePolicy2(igm),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupManagerBetaExists(
						"google_compute_instance_group_manager.igm-rolling-update-policy", &manager),
					resource.TestCheckResourceAttr(
						"google_compute_instance_group_manager.igm-rolling-update-policy", "update_strategy", "ROLLING_UPDATE"),
					resource.TestCheckResourceAttr(
						"google_compute_instance_group_manager.igm-rolling-update-policy", "rolling_update_policy.0.type", "PROACTIVE"),
					resource.TestCheckResourceAttr(
						"google_compute_instance_group_manager.igm-rolling-update-policy", "rolling_update_policy.0.minimal_action", "REPLACE"),
					resource.TestCheckResourceAttr(
						"google_compute_instance_group_manager.igm-rolling-update-policy", "rolling_update_policy.0.max_surge_fixed", "2"),
					resource.TestCheckResourceAttr(
						"google_compute_instance_group_manager.igm-rolling-update-policy", "rolling_update_policy.0.max_unavailable_fixed", "2"),
					resource.TestCheckResourceAttr(
						"google_compute_instance_group_manager.igm-rolling-update-policy", "rolling_update_policy.0.min_ready_sec", "20"),
					testAccCheckInstanceGroupManagerRollingUpdatePolicy(
						&manager, "google_compute_instance_group_manager.igm-rolling-update-policy"),
				),
			},
		},
	})
}

func TestAccInstanceGroupManager_separateRegions(t *testing.T) {
	t.Parallel()

	var manager compute.InstanceGroupManager

	igm1 := fmt.Sprintf("igm-test-%s", acctest.RandString(10))
	igm2 := fmt.Sprintf("igm-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceGroupManagerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceGroupManager_separateRegions(igm1, igm2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupManagerExists(
						"google_compute_instance_group_manager.igm-basic", &manager),
					testAccCheckInstanceGroupManagerExists(
						"google_compute_instance_group_manager.igm-basic-2", &manager),
				),
			},
		},
	})
}

func TestAccInstanceGroupManager_autoHealingPolicies(t *testing.T) {
	t.Parallel()

	var manager computeBeta.InstanceGroupManager

	template := fmt.Sprintf("igm-test-%s", acctest.RandString(10))
	target := fmt.Sprintf("igm-test-%s", acctest.RandString(10))
	igm := fmt.Sprintf("igm-test-%s", acctest.RandString(10))
	hck := fmt.Sprintf("igm-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceGroupManagerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceGroupManager_autoHealingPolicies(template, target, igm, hck),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceGroupManagerBetaExists(
						"google_compute_instance_group_manager.igm-basic", &manager),
					testAccCheckInstanceGroupManagerAutoHealingPolicies("google_compute_instance_group_manager.igm-basic", hck, 10),
				),
			},
			resource.TestStep{
				ResourceName:      "google_compute_instance_group_manager.igm-basic",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// This test is to make sure that a single version resource can link to a versioned resource
// without perpetual diffs because the self links mismatch.
// Once auto_healing_policies is no longer beta, we will need to use a new field or resource
// with Beta fields.
func TestAccInstanceGroupManager_selfLinkStability(t *testing.T) {
	t.Parallel()

	var manager computeBeta.InstanceGroupManager

	template := fmt.Sprintf("igm-test-%s", acctest.RandString(10))
	target := fmt.Sprintf("igm-test-%s", acctest.RandString(10))
	igm := fmt.Sprintf("igm-test-%s", acctest.RandString(10))
	hck := fmt.Sprintf("igm-test-%s", acctest.RandString(10))
	autoscaler := fmt.Sprintf("igm-test-%s", acctest.RandString(10))

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckInstanceGroupManagerDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccInstanceGroupManager_selfLinkStability(template, target, igm, hck, autoscaler),
				Check: testAccCheckInstanceGroupManagerBetaExists(
					"google_compute_instance_group_manager.igm-basic", &manager),
			},
		},
	})
}

func testAccCheckInstanceGroupManagerDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "google_compute_instance_group_manager" {
			continue
		}
		_, err := config.clientCompute.InstanceGroupManagers.Get(
			config.Project, rs.Primary.Attributes["zone"], rs.Primary.ID).Do()
		if err == nil {
			return fmt.Errorf("InstanceGroupManager still exists")
		}
	}

	return nil
}

func testAccCheckInstanceGroupManagerExists(n string, manager *compute.InstanceGroupManager) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.clientCompute.InstanceGroupManagers.Get(
			config.Project, rs.Primary.Attributes["zone"], rs.Primary.ID).Do()
		if err != nil {
			return err
		}

		if found.Name != rs.Primary.ID {
			return fmt.Errorf("InstanceGroupManager not found")
		}

		*manager = *found

		return nil
	}
}

func testAccCheckInstanceGroupManagerBetaExists(n string, manager *computeBeta.InstanceGroupManager) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		found, err := config.clientComputeBeta.InstanceGroupManagers.Get(
			config.Project, rs.Primary.Attributes["zone"], rs.Primary.ID).Do()
		if err != nil {
			return err
		}

		if found.Name != rs.Primary.ID {
			return fmt.Errorf("InstanceGroupManager not found")
		}

		*manager = *found

		return nil
	}
}

func testAccCheckInstanceGroupManagerUpdated(n string, size int64, targetPools []string, template string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		manager, err := config.clientCompute.InstanceGroupManagers.Get(
			config.Project, rs.Primary.Attributes["zone"], rs.Primary.ID).Do()
		if err != nil {
			return err
		}

		// Cannot check the target pool as the instance creation is asynchronous.  However, can
		// check the target_size.
		if manager.TargetSize != size {
			return fmt.Errorf("instance count incorrect")
		}

		tpNames := make([]string, 0, len(manager.TargetPools))
		for _, targetPool := range manager.TargetPools {
			tpNames = append(tpNames, GetResourceNameFromSelfLink(targetPool))
		}

		sort.Strings(tpNames)
		sort.Strings(targetPools)
		if !reflect.DeepEqual(tpNames, targetPools) {
			return fmt.Errorf("target pools incorrect. Expected %s, got %s", targetPools, tpNames)
		}

		// check that the instance template updated
		instanceTemplate, err := config.clientCompute.InstanceTemplates.Get(
			config.Project, template).Do()
		if err != nil {
			return fmt.Errorf("Error reading instance template: %s", err)
		}

		if instanceTemplate.Name != template {
			return fmt.Errorf("instance template not updated")
		}

		return nil
	}
}

func testAccCheckInstanceGroupManagerNamedPorts(n string, np map[string]int64, instanceGroupManager *compute.InstanceGroupManager) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		manager, err := config.clientCompute.InstanceGroupManagers.Get(
			config.Project, rs.Primary.Attributes["zone"], rs.Primary.ID).Do()
		if err != nil {
			return err
		}

		var found bool
		for _, namedPort := range manager.NamedPorts {
			found = false
			for name, port := range np {
				if namedPort.Name == name && namedPort.Port == port {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("named port incorrect")
			}
		}

		return nil
	}
}

func testAccCheckInstanceGroupManagerAutoHealingPolicies(n, hck string, initialDelaySec int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		manager, err := config.clientComputeBeta.InstanceGroupManagers.Get(
			config.Project, rs.Primary.Attributes["zone"], rs.Primary.ID).Do()
		if err != nil {
			return err
		}

		if len(manager.AutoHealingPolicies) != 1 {
			return fmt.Errorf("Expected # of auto healing policies to be 1, got %d", len(manager.AutoHealingPolicies))
		}
		autoHealingPolicy := manager.AutoHealingPolicies[0]

		if !strings.Contains(autoHealingPolicy.HealthCheck, hck) {
			return fmt.Errorf("Expected string \"%s\" to appear in \"%s\"", hck, autoHealingPolicy.HealthCheck)
		}

		if autoHealingPolicy.InitialDelaySec != initialDelaySec {
			return fmt.Errorf("Expected auto healing policy inital delay to be %d, got %d", initialDelaySec, autoHealingPolicy.InitialDelaySec)
		}
		return nil
	}
}

func testAccCheckInstanceGroupManagerTemplateTags(n string, tags []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)

		manager, err := config.clientCompute.InstanceGroupManagers.Get(
			config.Project, rs.Primary.Attributes["zone"], rs.Primary.ID).Do()
		if err != nil {
			return err
		}

		// check that the instance template updated
		instanceTemplate, err := config.clientCompute.InstanceTemplates.Get(
			config.Project, GetResourceNameFromSelfLink(manager.InstanceTemplate)).Do()
		if err != nil {
			return fmt.Errorf("Error reading instance template: %s", err)
		}

		if !reflect.DeepEqual(instanceTemplate.Properties.Tags.Items, tags) {
			return fmt.Errorf("instance template not updated")
		}

		return nil
	}
}

func testAccCheckInstanceGroupManagerUpdateStrategy(n, strategy string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		if rs.Primary.Attributes["update_strategy"] != strategy {
			return fmt.Errorf("Expected strategy to be %s, got %s",
				strategy, rs.Primary.Attributes["update_strategy"])
		}
		return nil
	}
}

func testAccCheckInstanceGroupManagerRollingUpdatePolicy(manager *computeBeta.InstanceGroupManager, resource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[resource]

		updatePolicy := manager.UpdatePolicy

		surgeFixed, _ := strconv.ParseInt(rs.Primary.Attributes["rolling_update_policy.0.max_surge_fixed"], 10, 64)
		if updatePolicy.MaxSurge.Fixed != surgeFixed {
			return fmt.Errorf("Expected update policy MaxSurge to be %d, got %d", surgeFixed, updatePolicy.MaxSurge.Fixed)
		}

		surgePercent, _ := strconv.ParseInt(rs.Primary.Attributes["rolling_update_policy.0.max_surge_percent"], 10, 64)
		if updatePolicy.MaxSurge.Percent != surgePercent {
			return fmt.Errorf("Expected update policy MaxSurge to be %d, got %d", surgePercent, updatePolicy.MaxSurge.Percent)
		}

		unavailableFixed, _ := strconv.ParseInt(rs.Primary.Attributes["rolling_update_policy.0.max_unavailable_fixed"], 10, 64)
		if updatePolicy.MaxUnavailable.Fixed != unavailableFixed {
			return fmt.Errorf("Expected update policy MaxUnavailable to be %d, got %d", unavailableFixed, updatePolicy.MaxUnavailable.Fixed)
		}

		unavailablePercent, _ := strconv.ParseInt(rs.Primary.Attributes["rolling_update_policy.0.max_unavailable_percent"], 10, 64)
		if updatePolicy.MaxUnavailable.Percent != unavailablePercent {
			return fmt.Errorf("Expected update policy MaxUnavailable to be %d, got %d", unavailablePercent, updatePolicy.MaxUnavailable.Percent)
		}

		policyType := rs.Primary.Attributes["rolling_update_policy.0.type"]
		if updatePolicy.Type != policyType {
			return fmt.Errorf("Expected  update policy Type to be  \"%s\", got \"%s\"", policyType, updatePolicy.Type)
		}

		policyAction := rs.Primary.Attributes["rolling_update_policy.0.minimal_action"]
		if updatePolicy.MinimalAction != policyAction {
			return fmt.Errorf("Expected  update policy MinimalAction to be  \"%s\", got \"%s\"", policyAction, updatePolicy.MinimalAction)
		}

		minReadySec, _ := strconv.ParseInt(rs.Primary.Attributes["rolling_update_policy.0.min_ready_sec"], 10, 64)
		if updatePolicy.MinReadySec != minReadySec {
			return fmt.Errorf("Expected update policy MinReadySec to be %d, got %d", minReadySec, updatePolicy.MinReadySec)
		}
		return nil
	}
}

func testAccInstanceGroupManager_basic(template, target, igm1, igm2 string) string {
	return fmt.Sprintf(`
	resource "google_compute_instance_template" "igm-basic" {
		name = "%s"
		machine_type = "n1-standard-1"
		can_ip_forward = false
		tags = ["foo", "bar"]

		disk {
			source_image = "debian-cloud/debian-8-jessie-v20160803"
			auto_delete = true
			boot = true
		}

		network_interface {
			network = "default"
		}

		metadata {
			foo = "bar"
		}

		service_account {
			scopes = ["userinfo-email", "compute-ro", "storage-ro"]
		}
	}

	resource "google_compute_target_pool" "igm-basic" {
		description = "Resource created for Terraform acceptance testing"
		name = "%s"
		session_affinity = "CLIENT_IP_PROTO"
	}

	resource "google_compute_instance_group_manager" "igm-basic" {
		description = "Terraform test instance group manager"
		name = "%s"
		instance_template = "${google_compute_instance_template.igm-basic.self_link}"
		target_pools = ["${google_compute_target_pool.igm-basic.self_link}"]
		base_instance_name = "igm-basic"
		zone = "us-central1-c"
		target_size = 2
	}

	resource "google_compute_instance_group_manager" "igm-no-tp" {
		description = "Terraform test instance group manager"
		name = "%s"
		instance_template = "${google_compute_instance_template.igm-basic.self_link}"
		base_instance_name = "igm-no-tp"
		zone = "us-central1-c"
		target_size = 2
	}
	`, template, target, igm1, igm2)
}

func testAccInstanceGroupManager_targetSizeZero(template, igm string) string {
	return fmt.Sprintf(`
	resource "google_compute_instance_template" "igm-basic" {
		name = "%s"
		machine_type = "n1-standard-1"
		can_ip_forward = false
		tags = ["foo", "bar"]

		disk {
			source_image = "debian-cloud/debian-8-jessie-v20160803"
			auto_delete = true
			boot = true
		}

		network_interface {
			network = "default"
		}

		metadata {
			foo = "bar"
		}

		service_account {
			scopes = ["userinfo-email", "compute-ro", "storage-ro"]
		}
	}

	resource "google_compute_instance_group_manager" "igm-basic" {
		description = "Terraform test instance group manager"
		name = "%s"
		instance_template = "${google_compute_instance_template.igm-basic.self_link}"
		base_instance_name = "igm-basic"
		zone = "us-central1-c"
	}
	`, template, igm)
}

func testAccInstanceGroupManager_update(template, target, igm string) string {
	return fmt.Sprintf(`
	resource "google_compute_instance_template" "igm-update" {
		name = "%s"
		machine_type = "n1-standard-1"
		can_ip_forward = false
		tags = ["foo", "bar"]

		disk {
			source_image = "debian-cloud/debian-8-jessie-v20160803"
			auto_delete = true
			boot = true
		}

		network_interface {
			network = "default"
		}

		metadata {
			foo = "bar"
		}

		service_account {
			scopes = ["userinfo-email", "compute-ro", "storage-ro"]
		}
	}

	resource "google_compute_target_pool" "igm-update" {
		description = "Resource created for Terraform acceptance testing"
		name = "%s"
		session_affinity = "CLIENT_IP_PROTO"
	}

	resource "google_compute_instance_group_manager" "igm-update" {
		description = "Terraform test instance group manager"
		name = "%s"
		instance_template = "${google_compute_instance_template.igm-update.self_link}"
		target_pools = ["${google_compute_target_pool.igm-update.self_link}"]
		base_instance_name = "igm-update"
		zone = "us-central1-c"
		target_size = 2
		named_port {
			name = "customhttp"
			port = 8080
		}
	}`, template, target, igm)
}

// Change IGM's instance template and target size
func testAccInstanceGroupManager_update2(template1, target1, target2, template2, igm string) string {
	return fmt.Sprintf(`
	resource "google_compute_instance_template" "igm-update" {
		name = "%s"
		machine_type = "n1-standard-1"
		can_ip_forward = false
		tags = ["foo", "bar"]

		disk {
			source_image = "debian-cloud/debian-8-jessie-v20160803"
			auto_delete = true
			boot = true
		}

		network_interface {
			network = "default"
		}

		metadata {
			foo = "bar"
		}

		service_account {
			scopes = ["userinfo-email", "compute-ro", "storage-ro"]
		}
	}

	resource "google_compute_target_pool" "igm-update" {
		description = "Resource created for Terraform acceptance testing"
		name = "%s"
		session_affinity = "CLIENT_IP_PROTO"
	}

	resource "google_compute_target_pool" "igm-update2" {
		description = "Resource created for Terraform acceptance testing"
		name = "%s"
		session_affinity = "CLIENT_IP_PROTO"
	}

	resource "google_compute_instance_template" "igm-update2" {
		name = "%s"
		machine_type = "n1-standard-1"
		can_ip_forward = false
		tags = ["foo", "bar"]

		disk {
			source_image = "debian-cloud/debian-8-jessie-v20160803"
			auto_delete = true
			boot = true
		}

		network_interface {
			network = "default"
		}

		metadata {
			foo = "bar"
		}

		service_account {
			scopes = ["userinfo-email", "compute-ro", "storage-ro"]
		}
	}

	resource "google_compute_instance_group_manager" "igm-update" {
		description = "Terraform test instance group manager"
		name = "%s"
		instance_template = "${google_compute_instance_template.igm-update2.self_link}"
		target_pools = [
			"${google_compute_target_pool.igm-update.self_link}",
			"${google_compute_target_pool.igm-update2.self_link}",
		]
		base_instance_name = "igm-update"
		zone = "us-central1-c"
		target_size = 3
		named_port {
			name = "customhttp"
			port = 8080
		}
		named_port {
			name = "customhttps"
			port = 8443
		}
	}`, template1, target1, target2, template2, igm)
}

func testAccInstanceGroupManager_updateLifecycle(tag, igm string) string {
	return fmt.Sprintf(`
	resource "google_compute_instance_template" "igm-update" {
		machine_type = "n1-standard-1"
		can_ip_forward = false
		tags = ["%s"]

		disk {
			source_image = "debian-cloud/debian-8-jessie-v20160803"
			auto_delete = true
			boot = true
		}

		network_interface {
			network = "default"
		}

		service_account {
			scopes = ["userinfo-email", "compute-ro", "storage-ro"]
		}

		lifecycle {
			create_before_destroy = true
		}
	}

	resource "google_compute_instance_group_manager" "igm-update" {
		description = "Terraform test instance group manager"
		name = "%s"
		instance_template = "${google_compute_instance_template.igm-update.self_link}"
		base_instance_name = "igm-update"
		zone = "us-central1-c"
		target_size = 2
		named_port {
			name = "customhttp"
			port = 8080
		}
	}`, tag, igm)
}

func testAccInstanceGroupManager_updateStrategy(igm string) string {
	return fmt.Sprintf(`
	resource "google_compute_instance_template" "igm-update-strategy" {
		machine_type = "n1-standard-1"
		can_ip_forward = false
		tags = ["terraform-testing"]

		disk {
			source_image = "debian-cloud/debian-8-jessie-v20160803"
			auto_delete = true
			boot = true
		}

		network_interface {
			network = "default"
		}

		service_account {
			scopes = ["userinfo-email", "compute-ro", "storage-ro"]
		}

		lifecycle {
			create_before_destroy = true
		}
	}

	resource "google_compute_instance_group_manager" "igm-update-strategy" {
		description = "Terraform test instance group manager"
		name = "%s"
		instance_template = "${google_compute_instance_template.igm-update-strategy.self_link}"
		base_instance_name = "igm-update-strategy"
		zone = "us-central1-c"
		target_size = 2
		update_strategy = "NONE"
		named_port {
			name = "customhttp"
			port = 8080
		}
	}`, igm)
}

func testAccInstanceGroupManager_rollingUpdatePolicy(igm string) string {
	return fmt.Sprintf(`
resource "google_compute_instance_template" "igm-rolling-update-policy" {
	machine_type = "n1-standard-1"
	can_ip_forward = false
	tags = ["terraform-testing"]

	disk {
		source_image = "debian-cloud/debian-8-jessie-v20160803"
		auto_delete = true
		boot = true
	}

	network_interface {
		network = "default"
	}

	service_account {
		scopes = ["userinfo-email", "compute-ro", "storage-ro"]
	}

	lifecycle {
		create_before_destroy = true
	}
}

resource "google_compute_instance_group_manager" "igm-rolling-update-policy" {
	description = "Terraform test instance group manager"
	name = "%s"
	instance_template = "${google_compute_instance_template.igm-rolling-update-policy.self_link}"
	base_instance_name = "igm-rolling-update-policy"
	zone = "us-central1-c"
	target_size = 3
	update_strategy = "ROLLING_UPDATE"
	rolling_update_policy {
		type = "PROACTIVE"
		minimal_action = "REPLACE"
		max_surge_percent = 50
		max_unavailable_percent = 50
		min_ready_sec = 20
	}
	named_port {
		name = "customhttp"
		port = 8080
	}
}`, igm)
}

func testAccInstanceGroupManager_rollingUpdatePolicy2(igm string) string {
	return fmt.Sprintf(`
resource "google_compute_instance_template" "igm-rolling-update-policy" {
	machine_type = "n1-standard-1"
	can_ip_forward = false
	tags = ["terraform-testing"]

	disk {
		source_image = "debian-cloud/debian-8-jessie-v20160803"
		auto_delete = true
		boot = true
	}

	network_interface {
		network = "default"
	}

	lifecycle {
		create_before_destroy = true
	}
}

resource "google_compute_instance_group_manager" "igm-rolling-update-policy" {
	description = "Terraform test instance group manager"
	name = "%s"
	instance_template = "${google_compute_instance_template.igm-rolling-update-policy.self_link}"
	base_instance_name = "igm-rolling-update-policy"
	zone = "us-central1-c"
	target_size = 3
	update_strategy = "ROLLING_UPDATE"
	rolling_update_policy {
		type = "PROACTIVE"
		minimal_action = "REPLACE"
		max_surge_fixed = 2
		max_unavailable_fixed = 2
		min_ready_sec = 20
	}
	named_port {
		name = "customhttp"
		port = 8080
	}
}`, igm)
}

func testAccInstanceGroupManager_separateRegions(igm1, igm2 string) string {
	return fmt.Sprintf(`
	resource "google_compute_instance_template" "igm-basic" {
		machine_type = "n1-standard-1"
		can_ip_forward = false
		tags = ["foo", "bar"]

		disk {
			source_image = "debian-cloud/debian-8-jessie-v20160803"
			auto_delete = true
			boot = true
		}

		network_interface {
			network = "default"
		}

		metadata {
			foo = "bar"
		}

		service_account {
			scopes = ["userinfo-email", "compute-ro", "storage-ro"]
		}
	}

	resource "google_compute_instance_group_manager" "igm-basic" {
		description = "Terraform test instance group manager"
		name = "%s"
		instance_template = "${google_compute_instance_template.igm-basic.self_link}"
		base_instance_name = "igm-basic"
		zone = "us-central1-c"
		target_size = 2
	}

	resource "google_compute_instance_group_manager" "igm-basic-2" {
		description = "Terraform test instance group manager"
		name = "%s"
		instance_template = "${google_compute_instance_template.igm-basic.self_link}"
		base_instance_name = "igm-basic-2"
		zone = "us-west1-b"
		target_size = 2
	}
	`, igm1, igm2)
}

func testAccInstanceGroupManager_autoHealingPolicies(template, target, igm, hck string) string {
	return fmt.Sprintf(`
resource "google_compute_instance_template" "igm-basic" {
	name = "%s"
	machine_type = "n1-standard-1"
	can_ip_forward = false
	tags = ["foo", "bar"]
	disk {
		source_image = "debian-cloud/debian-8-jessie-v20160803"
		auto_delete = true
		boot = true
	}
	network_interface {
		network = "default"
	}
	metadata {
		foo = "bar"
	}
	service_account {
		scopes = ["userinfo-email", "compute-ro", "storage-ro"]
	}
}

resource "google_compute_target_pool" "igm-basic" {
	description = "Resource created for Terraform acceptance testing"
	name = "%s"
	session_affinity = "CLIENT_IP_PROTO"
}

resource "google_compute_instance_group_manager" "igm-basic" {
	description = "Terraform test instance group manager"
	name = "%s"
	instance_template = "${google_compute_instance_template.igm-basic.self_link}"
	target_pools = ["${google_compute_target_pool.igm-basic.self_link}"]
	base_instance_name = "igm-basic"
	zone = "us-central1-c"
	target_size = 2
	auto_healing_policies {
		health_check = "${google_compute_http_health_check.zero.self_link}"
		initial_delay_sec = "10"
	}
}

resource "google_compute_http_health_check" "zero" {
	name               = "%s"
	request_path       = "/"
	check_interval_sec = 1
	timeout_sec        = 1
}
	`, template, target, igm, hck)
}

// This test is to make sure that a single version resource can link to a versioned resource
// without perpetual diffs because the self links mismatch.
// Once auto_healing_policies is no longer beta, we will need to use a new field or resource
// with Beta fields.
func testAccInstanceGroupManager_selfLinkStability(template, target, igm, hck, autoscaler string) string {
	return fmt.Sprintf(`
resource "google_compute_instance_template" "igm-basic" {
	name = "%s"
	machine_type = "n1-standard-1"
	can_ip_forward = false
	tags = ["foo", "bar"]
	disk {
		source_image = "debian-cloud/debian-8-jessie-v20160803"
		auto_delete = true
		boot = true
	}
	network_interface {
		network = "default"
	}
	metadata {
		foo = "bar"
	}
	service_account {
		scopes = ["userinfo-email", "compute-ro", "storage-ro"]
	}
}

resource "google_compute_target_pool" "igm-basic" {
	description = "Resource created for Terraform acceptance testing"
	name = "%s"
	session_affinity = "CLIENT_IP_PROTO"
}

resource "google_compute_instance_group_manager" "igm-basic" {
	description = "Terraform test instance group manager"
	name = "%s"
	instance_template = "${google_compute_instance_template.igm-basic.self_link}"
	target_pools = ["${google_compute_target_pool.igm-basic.self_link}"]
	base_instance_name = "igm-basic"
	zone = "us-central1-c"
	target_size = 2
	auto_healing_policies {
		health_check = "${google_compute_http_health_check.zero.self_link}"
		initial_delay_sec = "10"
	}
}

resource "google_compute_http_health_check" "zero" {
	name               = "%s"
	request_path       = "/"
	check_interval_sec = 1
	timeout_sec        = 1
}

resource "google_compute_autoscaler" "foobar" {
	name = "%s"
	zone = "us-central1-c"
	target = "${google_compute_instance_group_manager.igm-basic.self_link}"
	autoscaling_policy = {
		max_replicas = 10
		min_replicas = 1
		cooldown_period = 60
		cpu_utilization = {
			target = 0.5
		}
	}
}
`, template, target, igm, hck, autoscaler)
}
