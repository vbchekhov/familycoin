package main

import (
	"testing"
	"time"
)

func Test_Group(t *testing.T) {
	type args struct {
		user  *User
		start time.Time
		end   time.Time
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "debit 1",
			args: args{
				user: &User{
					TelegramId: 256674624,
					FullName:   "",
					FamilyId:   1,
				},
				start: time.Now().Add(-time.Hour * 24 * 365 * 20),
				end:   time.Now(),
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Details
			if got = Group(&Debit{}, tt.args.user.TelegramId, tt.args.start, tt.args.end); len(got) == tt.want {
				t.Errorf("Group() = %v, want %v", got, tt.want)
			}
			t.Logf("%+v", got)
		})
	}
}

func Test_Detail(t *testing.T) {
	type args struct {
		user  *User
		start time.Time
		end   time.Time
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "debit 1",
			args: args{
				user: &User{
					TelegramId: 256674624,
					FullName:   "",
					FamilyId:   1,
				},
				start: time.Now().Add(-time.Hour * 24 * 365 * 20),
				end:   time.Now(),
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Details
			if got = Detail(&Debit{}, tt.args.user.TelegramId, tt.args.start, tt.args.end); len(got) == tt.want {
				t.Errorf("Detail() = %v, want %v", got, tt.want)
			}
			t.Logf("%+v", got)
		})
	}
}

func TestBalance(t *testing.T) {

	t.Logf("%+v", GetBalance(256674624))

}

func TestUser_family(t *testing.T) {

	u := &User{TelegramId: 256674624}
	u.read()
	t.Log(u.Family())
}

func TestGetPeggyBank(t *testing.T) {

	now := time.Now()

	weeks := []int{0, 1, 2, 3, 4}
	for i := range weeks {

		year, week := now.Add(-time.Hour * 24 * 7 * time.Duration(weeks[i])).ISOWeek()
		peggy, err := GetPeggyBank(256674624, week, year)
		if err != nil {
			t.Fatal(err)
		}

		// start, end := DaysOfISOWeek(year, week, now.Location())
		t.Logf("%+v %d %d start/end %s/%s", peggy, week, year, peggy.Monday.Format("2006-02-01"), peggy.Sunday.Format("2006-02-01"))

	}

}
