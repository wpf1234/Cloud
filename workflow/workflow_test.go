package workflow

import (
	"fmt"
	"testing"
)

func TestWorkflow_GetToken(t *testing.T) {
	type args struct {
		username string
		passwd   string
	}

	tests := []struct {
		name    string
		w       *Workflow
		args    args
		want    string
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "",
			w:    nil,
			args: args{
				username: "180435",
				passwd:   "wpf!2345",
			},
			want:    "",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Workflow{}
			got, err := w.GetToken(tt.args.username, tt.args.passwd)
			if (err != nil) != tt.wantErr {
				t.Errorf("Workflow.GetToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Workflow.GetToken() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestWorkflow_ActiveWorkflow(t *testing.T) {
	wf := NewWorkflowClient()
	tk, err := wf.GetToken(`180435`, `wpf!2345`)
	if err != nil {
		fmt.Println(err)
	}

	type args struct {
		WFID              string
		WFNodeOperationID string
		MainData          string
		token             string
	}

	tests := []struct {
		name    string
		w       *Workflow
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "提交软件服务发布申请",
			w:    nil,
			args: args{
				WFID:              "5157",
				WFNodeOperationID: "24604",
				MainData:          `{"邮箱号": "180435", "开发者（姓名）":"阿百川"}`,
				token:             tk,
			},
			wantErr: false,
		},
		// {
		//  name: "网络",
		//  w:    nil,
		//  args: args{
		//     WFID:              "4868",
		//     WFNodeOperationID: "22807",
		//     MainData:          `{"邮箱": "180256"}`,
		//     token:             tk,
		//  },
		//  wantErr: false,
		// },
		// {
		//  name: "系统",
		//  w:    nil,
		//  args: args{
		//     WFID:              "4881",
		//     WFNodeOperationID: "22858",
		//     MainData:          `{"离职人员邮箱号": "180256","离职人员姓名": "綦麟","部门": "计算机中心"}`,
		//     token:             tk,
		//  },
		//  wantErr: false,
		// },
		// {
		//  name: "安全",
		//  w:    nil,
		//  args: args{
		//     WFID:              "4866",
		//     WFNodeOperationID: "22791",
		//     MainData:          `{"离职人员邮箱号": "180256","离职人员姓名": "綦麟","部门": "计算机中心"}`,
		//     token:             tk,
		//  },
		//  wantErr: false,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Workflow{}
			if err := w.ActiveWorkflow(tt.args.WFID, tt.args.WFNodeOperationID, tt.args.MainData, `{}`, tt.args.token); (err != nil) != tt.wantErr {
				t.Errorf("Workflow.ActiveWorkflow() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
