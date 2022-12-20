package stardog

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_IsAlive_true(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/admin/alive", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusOK)
	})
	ctx := context.Background()
	got, _, err := client.ServerAdmin.IsAlive(ctx)
	if err != nil {
		t.Errorf("ServerAdmin.IsAlive returned error: %v", err)
	}
	if want := true; !cmp.Equal(*got, want) {
		t.Errorf("ServerAdmin.IsAlive = %+v, want %+v", *got, want)
	}

	const methodName = "IsAlive"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.ServerAdmin.IsAlive(nil)
		if got != nil && *got != false {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want false", methodName, *got)
		}
		return resp, err
	})
}

func Test_IsAlive_false(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	mux.HandleFunc("/admin/alive", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		w.WriteHeader(http.StatusServiceUnavailable)
	})
	ctx := context.Background()
	got, _, err := client.ServerAdmin.IsAlive(ctx)
	if err == nil {
		t.Errorf("ServerAdmin.IsAlive returned nil error")
	}
	if want := false; !cmp.Equal(*got, want) {
		t.Errorf("ServerAdmin.IsAlive = %+v, want %+v", *got, want)
	}

	const methodName = "IsAlive"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.ServerAdmin.IsAlive(nil)
		if got != nil && *got != false {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want false", methodName, *got)
		}
		return resp, err
	})
}

func Test_GetProcesses(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var processesJSON = `[
  {
    "type": "Transaction",
    "kernelId": "3d6d135c-6b12-48c8-aa22-4f955fa7bea9",
    "id": "c273226b-de41-407d-9343-6157cfbbedb1",
    "db": "myDb",
    "user": "noah.gorstein@stardog.com",
    "startTime": 1669949829376,
    "status": "RUNNING",
    "progress": {
      "max": 0,
      "current": 0,
      "stage": ""
    }
  }
]`
	var wantProcesses = &[]Process{
		{
			Type:      "Transaction",
			KernelID:  "3d6d135c-6b12-48c8-aa22-4f955fa7bea9",
			ID:        "c273226b-de41-407d-9343-6157cfbbedb1",
			Db:        "myDb",
			User:      "noah.gorstein@stardog.com",
			StartTime: 1669949829376,
			Status:    "RUNNING",
			Progress:  ProcessProgress{Max: 0, Current: 0, Stage: ""},
		}}
	mux.HandleFunc("/admin/processes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(processesJSON))
	})

	ctx := context.Background()
	got, _, err := client.ServerAdmin.GetProcesses(ctx)
	if err != nil {
		t.Errorf("ServerAdmin.GetProcesses returned error: %v", err)
	}
	if want := wantProcesses; !cmp.Equal(got, want) {
		t.Errorf("ServerAdmin.GetProcesses = %+v, want %+v", got, want)
	}

	const methodName = "GetProcesses"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.ServerAdmin.GetProcesses(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_GetProcess(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	processID := "c273226b-de41-407d-9343-6157cfbbedb1"

	processJSON := `{
    "type": "Transaction",
    "kernelId": "3d6d135c-6b12-48c8-aa22-4f955fa7bea9",
    "id": "c273226b-de41-407d-9343-6157cfbbedb1",
    "db": "myDb",
    "user": "noah.gorstein@stardog.com",
    "startTime": 1669949829376,
    "status": "RUNNING",
    "progress": {
      "max": 0,
      "current": 0,
      "stage": ""
    }
  }
  `
	wantProcesses := &Process{
		Type:      "Transaction",
		KernelID:  "3d6d135c-6b12-48c8-aa22-4f955fa7bea9",
		ID:        "c273226b-de41-407d-9343-6157cfbbedb1",
		Db:        "myDb",
		User:      "noah.gorstein@stardog.com",
		StartTime: 1669949829376,
		Status:    "RUNNING",
		Progress:  ProcessProgress{Max: 0, Current: 0, Stage: ""},
	}
	mux.HandleFunc(fmt.Sprintf("/admin/processes/%s", processID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(processJSON))
	})

	ctx := context.Background()
	got, _, err := client.ServerAdmin.GetProcess(ctx, processID)
	if err != nil {
		t.Errorf("ServerAdmin.GetProcess returned error: %v", err)
	}
	if want := wantProcesses; !cmp.Equal(got, want) {
		t.Errorf("ServerAdmin.GetProcess = %+v, want %+v", got, want)
	}

	const methodName = "GetProcess"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.ServerAdmin.GetProcess(nil, processID)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_KillProcess(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	processID := "c273226b-de41-407d-9343-6157cfbbedb1"
	mux.HandleFunc(fmt.Sprintf("/admin/processes/%s", processID), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")

		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	_, err := client.ServerAdmin.KillProcess(ctx, processID)
	if err != nil {
		t.Errorf("Security.DeleteUser returned error: %v", err)
	}

	const methodName = "KillProcess"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.ServerAdmin.KillProcess(nil, processID)
	})
}
