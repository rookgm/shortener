package client

import (
	"testing"
)

func Test_authToken_Verify(t *testing.T) {

	tokenStr := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTA4NjkzNjUsInVzZXJfaWQiOiI1MWRiNGViYi04ZGFmLTQyMmYtYmQ4Ni03ZmFmYTNiN2YwMmEifQ.VXQh2HyzQXiS6TOKe1J_g9nPTs9QMCl5_UFwMsCtv8Y"
	key := "secretkey"

	tokenStrNoID := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTA4NzA1NDJ9.5kAcIZ9DyU6767My_Q7B1ZDyL0laf1LDyG5xhPUY9Sc"

	type fields struct {
		secretKey []byte
	}
	type args struct {
		tokenString string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "valid_key",
			fields:  fields{[]byte(key)},
			args:    args{tokenStr},
			want:    "51db4ebb-8daf-422f-bd86-7fafa3b7f02a",
			wantErr: false,
		},
		{
			name:    "bad_key",
			fields:  fields{[]byte("badkey")},
			args:    args{tokenStr},
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty_token",
			fields:  fields{[]byte("empty")},
			args:    args{""},
			want:    "",
			wantErr: true,
		},
		{
			name:    "claim_without_id",
			fields:  fields{[]byte(key)},
			args:    args{tokenStrNoID},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &authToken{
				secretKey: tt.fields.secretKey,
			}
			got, err := a.Verify(tt.args.tokenString)
			if (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Verify() got = %v, want %v", got, tt.want)
			}
		})
	}
}
