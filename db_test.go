package main

import (
	"database/sql"
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

	var res []struct {
		Balance int `gorm:"column:b"`
	}

	u := &User{TelegramId: 256674624}
	u.read()

	db.Exec(`
	create or replace table balance (
		select
			   sum(d.sum) as debit
		from debits as d
				 left join debit_types dt on d.debit_type_id = dt.id
				 left join users u on u.id = d.user_id
		where d.user_id in (
			select distinct id
			from users
			where users.family_id = @family_id or users.telegram_id = @telegram_id)
	
		union all
	
		select
			   sum(-c.sum) as debit
		from credits as c
				 left join credit_types ct on c.credit_type_id = ct.id
				 left join users u on u.id = c.user_id
		where c.user_id in (
			select distinct id
			from users
			where users.family_id = @family_id or users.telegram_id = @telegram_id)
	);
	
	`, sql.Named("family_id", u.FamilyId), sql.Named("telegram_id", u.TelegramId)).
		Raw(`select sum(debit) as b from balance;`).Scan(&res)

	// r.Raw(`select sum(debit) from balance;`).Scan(&res)

	// db.Exec(`drop table balance;`)

	t.Logf("%+v", res)

}

func TestUser_family(t *testing.T) {

	u := &User{TelegramId: 256674624}
	u.read()
	t.Log(u.Family())
}
