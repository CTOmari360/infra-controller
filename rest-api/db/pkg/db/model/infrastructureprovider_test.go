// SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"context"
	"testing"

	cutil "github.com/NVIDIA/infra-controller/rest-api/common/pkg/util"
	"github.com/NVIDIA/infra-controller/rest-api/db/pkg/db"
	stracer "github.com/NVIDIA/infra-controller/rest-api/db/pkg/tracer"
	"github.com/NVIDIA/infra-controller/rest-api/db/pkg/util"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	otrace "go.opentelemetry.io/otel/trace"
)

func TestInfrastructureProviderSQLDAO_GetByID(t *testing.T) {
	type fields struct {
		dbSession *db.Session
	}
	type args struct {
		ctx context.Context
		id  uuid.UUID
	}

	// Create test DB
	dbSession := util.GetTestDBSession(t, false)

	// Create Infrastructure Provider table
	err := dbSession.DB.ResetModel(context.Background(), (*InfrastructureProvider)(nil))
	if err != nil {
		t.Fatal(err)
	}

	ip := &InfrastructureProvider{
		ID:             uuid.New(),
		Name:           "test",
		DisplayName:    cutil.GetPtr("test"),
		Org:            "test-org",
		OrgDisplayName: cutil.GetPtr("Test Org"),
		CreatedBy:      uuid.New(),
	}

	_, err = dbSession.DB.NewInsert().Model(ip).Exec(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// OTEL Spanner configuration
	ctx := context.Background()
	_, _, ctx = testCommonTraceProviderSetup(t, ctx)

	tests := []struct {
		name               string
		fields             fields
		args               args
		want               *InfrastructureProvider
		wantErr            bool
		wantErrVal         error
		verifyChildSpanner bool
	}{
		{
			name: "retrieve an InfrastructureProvider by ID",
			fields: fields{
				dbSession: dbSession,
			},
			args: args{
				ctx: ctx,
				id:  ip.ID,
			},
			want:    ip,
			wantErr: false,
		},
		{
			name: "error retrieve an InfrastructureProvider by ID",
			fields: fields{
				dbSession: dbSession,
			},
			args: args{
				ctx: context.Background(),
				id:  uuid.New(),
			},
			want:       nil,
			wantErr:    true,
			wantErrVal: db.ErrDoesNotExist,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipsd := InfrastructureProviderSQLDAO{
				dbSession: tt.fields.dbSession,
			}
			got, err := ipsd.GetByID(tt.args.ctx, nil, tt.args.id, nil)
			if !tt.wantErr {
				require.NoError(t, err)
			} else {
				assert.Equal(t, tt.wantErrVal, err)
				return
			}

			assert.Equal(t, tt.want.ID, got.ID)
			assert.Equal(t, tt.want.Name, got.Name)
			assert.Equal(t, *tt.want.DisplayName, *got.DisplayName)
			assert.Equal(t, tt.want.Org, got.Org)
			assert.Equal(t, *tt.want.OrgDisplayName, *got.OrgDisplayName)

			if tt.verifyChildSpanner {
				span := otrace.SpanFromContext(ctx)
				assert.True(t, span.SpanContext().IsValid())
				_, ok := ctx.Value(stracer.TracerKey).(otrace.Tracer)
				assert.True(t, ok)
			}
		})
	}
}

func TestInfrastructureProviderSQLDAO_GetAllByOrg(t *testing.T) {
	type fields struct {
		dbSession *db.Session
	}
	type args struct {
		ctx context.Context
		org string
	}

	// Create test DB
	dbSession := util.GetTestDBSession(t, false)
	defer dbSession.Close()

	// Create Infrastructure Provider table
	err := dbSession.DB.ResetModel(context.Background(), (*InfrastructureProvider)(nil))
	if err != nil {
		t.Fatal(err)
	}

	org := "test-org"
	orgDisplayName := "Test Org"

	ip1 := InfrastructureProvider{
		ID:             uuid.New(),
		Name:           "test 1",
		DisplayName:    cutil.GetPtr("test 2"),
		Org:            org,
		OrgDisplayName: cutil.GetPtr(orgDisplayName),
		CreatedBy:      uuid.New(),
	}

	ip2 := InfrastructureProvider{
		ID:             uuid.New(),
		Name:           "test 2",
		DisplayName:    cutil.GetPtr("test 2"),
		Org:            org,
		OrgDisplayName: cutil.GetPtr(orgDisplayName),
		CreatedBy:      uuid.New(),
	}

	ips := []InfrastructureProvider{ip1, ip2}

	_, err = dbSession.DB.NewInsert().Model(&ips).Exec(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// OTEL Spanner configuration
	ctx := context.Background()
	_, _, ctx = testCommonTraceProviderSetup(t, ctx)

	tests := []struct {
		name               string
		fields             fields
		args               args
		want               []InfrastructureProvider
		wantErr            bool
		verifyChildSpanner bool
	}{
		{
			name: "retrieve all InfrastructureProvider by org ID",
			fields: fields{
				dbSession: dbSession,
			},
			args: args{
				ctx: ctx,
				org: org,
			},
			want:               ips,
			wantErr:            false,
			verifyChildSpanner: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipsd := InfrastructureProviderSQLDAO{
				dbSession: tt.fields.dbSession,
			}
			got, err := ipsd.GetAllByOrg(tt.args.ctx, nil, tt.args.org, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("InfrastructureProviderSQLDAO.GetAllByOrg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("InfrastructureProviderSQLDAO.GetAllByOrg() gotlen = %v, wantlen = %v", len(got), len(tt.want))
			}

			if tt.verifyChildSpanner {
				span := otrace.SpanFromContext(ctx)
				assert.True(t, span.SpanContext().IsValid())
				_, ok := ctx.Value(stracer.TracerKey).(otrace.Tracer)
				assert.True(t, ok)
			}
		})
	}
}

func TestInfrastructureProviderSQLDAO_Create(t *testing.T) {
	type fields struct {
		dbSession *db.Session
	}
	type args struct {
		ctx   context.Context
		input InfrastructureProviderCreateInput
	}

	// Create test DB
	dbSession := util.GetTestDBSession(t, false)
	defer dbSession.Close()

	// Create Infrastructure Provider table
	err := dbSession.DB.ResetModel(context.Background(), (*InfrastructureProvider)(nil))
	if err != nil {
		t.Fatal(err)
	}

	ip := &InfrastructureProvider{
		Name:           "test",
		DisplayName:    cutil.GetPtr("test"),
		Org:            "test-org",
		OrgDisplayName: cutil.GetPtr("Test Org"),
		CreatedBy:      uuid.New(),
	}

	// OTEL Spanner configuration
	ctx := context.Background()
	_, _, ctx = testCommonTraceProviderSetup(t, ctx)

	tests := []struct {
		name               string
		fields             fields
		args               args
		want               *InfrastructureProvider
		wantErr            bool
		verifyChildSpanner bool
	}{
		{
			name: "create an InfrastructureProvider",
			fields: fields{
				dbSession: dbSession,
			},
			args: args{
				ctx: ctx,
				input: InfrastructureProviderCreateInput{
					Name:           ip.Name,
					DisplayName:    ip.DisplayName,
					Org:            ip.Org,
					OrgDisplayName: ip.OrgDisplayName,
					CreatedBy:      ip.CreatedBy,
				},
			},
			want:               ip,
			wantErr:            false,
			verifyChildSpanner: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipsd := InfrastructureProviderSQLDAO{
				dbSession: tt.fields.dbSession,
			}
			got, err := ipsd.Create(tt.args.ctx, nil, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("InfrastructureProviderSQLDAO.Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.want.Name)
			}

			if *got.DisplayName != *tt.want.DisplayName {
				t.Errorf("DisplayName = %v, want %v", *got.DisplayName, *tt.want.DisplayName)
			}

			if got.Org != tt.want.Org {
				t.Errorf("Org = %v, want %v", got.Org, tt.want.Org)
			}

			if got.OrgDisplayName != nil && *got.OrgDisplayName != *tt.want.OrgDisplayName {
				t.Errorf("OrgDisplayName = %v, want %v", *got.OrgDisplayName, *tt.want.OrgDisplayName)
			}

			if got.CreatedBy != tt.want.CreatedBy {
				t.Errorf("CreatedBy = %v, want %v", got.CreatedBy, tt.want.CreatedBy)
			}

			if tt.verifyChildSpanner {
				span := otrace.SpanFromContext(ctx)
				assert.True(t, span.SpanContext().IsValid())
				_, ok := ctx.Value(stracer.TracerKey).(otrace.Tracer)
				assert.True(t, ok)
			}
		})
	}
}

func TestInfrastructureProviderSQLDAO_Update(t *testing.T) {
	type fields struct {
		dbSession *db.Session
	}
	type args struct {
		ctx   context.Context
		input InfrastructureProviderUpdateInput
	}

	// Create test DB
	dbSession := util.GetTestDBSession(t, false)
	defer dbSession.Close()

	// Create Infrastructure Provider table
	err := dbSession.DB.ResetModel(context.Background(), (*InfrastructureProvider)(nil))
	if err != nil {
		t.Fatal(err)
	}

	// Create infrastructure provider
	ip := &InfrastructureProvider{
		ID:             uuid.New(),
		Name:           "test",
		DisplayName:    cutil.GetPtr("Test"),
		Org:            "test-org",
		OrgDisplayName: cutil.GetPtr("Test Org"),
		CreatedBy:      uuid.New(),
	}

	_, err = dbSession.DB.NewInsert().Model(ip).Exec(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// Updated infrastructure provider
	uip := &InfrastructureProvider{
		ID:             ip.ID,
		Name:           "test2",
		DisplayName:    cutil.GetPtr("Test 2"),
		Org:            ip.Org,
		OrgDisplayName: cutil.GetPtr("Test Org Updated"),
		CreatedBy:      ip.CreatedBy,
	}

	// OTEL Spanner configuration
	ctx := context.Background()
	_, _, ctx = testCommonTraceProviderSetup(t, ctx)

	tests := []struct {
		name               string
		fields             fields
		args               args
		want               *InfrastructureProvider
		wantErr            bool
		verifyChildSpanner bool
	}{
		{
			name: "update an InfrastructureProvider",
			fields: fields{
				dbSession: dbSession,
			},
			args: args{
				ctx: ctx,
				input: InfrastructureProviderUpdateInput{
					InfrastructureProviderID: ip.ID,
					Name:                     cutil.GetPtr(uip.Name),
					DisplayName:              uip.DisplayName,
					OrgDisplayName:           uip.OrgDisplayName,
				},
			},
			want:               uip,
			wantErr:            false,
			verifyChildSpanner: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipsd := InfrastructureProviderSQLDAO{
				dbSession: tt.fields.dbSession,
			}
			got, err := ipsd.Update(tt.args.ctx, nil, tt.args.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("InfrastructureProviderSQLDAO.Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got.Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.want.Name)
			}

			if *got.DisplayName != *tt.want.DisplayName {
				t.Errorf("DisplayName = %v, want %v", *got.DisplayName, *tt.want.DisplayName)
			}

			if got.Org != tt.want.Org {
				t.Errorf("Org = %v, want %v", got.Org, tt.want.Org)
			}

			if got.Updated.String() == tt.want.Updated.String() {
				t.Errorf("got.Updated = %v, want different value", got.Updated)
			}

			if tt.verifyChildSpanner {
				span := otrace.SpanFromContext(ctx)
				assert.True(t, span.SpanContext().IsValid())
				_, ok := ctx.Value(stracer.TracerKey).(otrace.Tracer)
				assert.True(t, ok)
			}
		})
	}
}

func TestInfrastructureProviderSQLDAO_Delete(t *testing.T) {
	type fields struct {
		dbSession *db.Session
	}
	type args struct {
		ctx context.Context
		id  uuid.UUID
	}

	// Create test DB
	dbSession := util.GetTestDBSession(t, false)
	defer dbSession.Close()

	// Create Infrastructure Provider table
	err := dbSession.DB.ResetModel(context.Background(), (*InfrastructureProvider)(nil))
	if err != nil {
		t.Fatal(err)
	}

	ip := &InfrastructureProvider{
		ID:             uuid.New(),
		Name:           "test",
		DisplayName:    cutil.GetPtr("test"),
		Org:            "test-org",
		OrgDisplayName: cutil.GetPtr("Test Org"),
	}

	_, err = dbSession.DB.NewInsert().Model(ip).Exec(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	// OTEL Spanner configuration
	ctx := context.Background()
	_, _, ctx = testCommonTraceProviderSetup(t, ctx)

	tests := []struct {
		name               string
		fields             fields
		args               args
		wantErr            bool
		verifyChildSpanner bool
	}{
		{
			name: "delete InfrastructureProvider by ID",
			fields: fields{
				dbSession: dbSession,
			},
			args: args{
				ctx: ctx,
				id:  ip.ID,
			},
			wantErr:            false,
			verifyChildSpanner: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ipsd := InfrastructureProviderSQLDAO{
				dbSession: tt.fields.dbSession,
			}
			if err := ipsd.Delete(tt.args.ctx, nil, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("InfrastructureProviderSQLDAO.Delete() error = %v, wantErr %v", err, tt.wantErr)
			}

			dip := &InfrastructureProvider{}
			err := dbSession.DB.NewSelect().Model(dip).WhereDeleted().Where("id = ?", ip.ID).Scan(context.Background())
			if err != nil {
				t.Fatal(err)
			}

			if dip.Deleted == nil {
				t.Errorf("Failed to soft-delete InfrastructureProvider")
			}

			if tt.verifyChildSpanner {
				span := otrace.SpanFromContext(ctx)
				assert.True(t, span.SpanContext().IsValid())
				_, ok := ctx.Value(stracer.TracerKey).(otrace.Tracer)
				assert.True(t, ok)
			}
		})
	}
}
