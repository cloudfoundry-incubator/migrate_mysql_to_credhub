package migrator_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/brokerapi"

	"code.cloudfoundry.org/lager/lagertest"
	"code.cloudfoundry.org/migrate_mysql_to_credhub/migrator"
	"code.cloudfoundry.org/migrate_mysql_to_credhub/migrator/fakes"
	"code.cloudfoundry.org/service-broker-store/brokerstore"
	"code.cloudfoundry.org/service-broker-store/brokerstore/brokerstorefakes"
)

var _ = Describe("Migrator", func() {
	var (
		migrationObj migrator.Migrator
		fromStore    *fakes.FakeRetirableStore
		toStore      *brokerstorefakes.FakeStore
	)

	BeforeEach(func() {
		logger := lagertest.NewTestLogger("migrator-test")
		migrationObj = migrator.NewMigrator(logger)
		fromStore = &fakes.FakeRetirableStore{}
		toStore = &brokerstorefakes.FakeStore{}
	})

	Context("when there are instance details in fromStore", func() {
		BeforeEach(func() {
			fromStore.RetrieveAllInstanceDetailsReturns(map[string]brokerstore.ServiceInstance{
				"123": brokerstore.ServiceInstance{ServiceID: "some-service-1"},
				"456": brokerstore.ServiceInstance{ServiceID: "some-service-2"},
			}, nil)
			err := migrationObj.Migrate(fromStore, toStore)
			Expect(err).NotTo(HaveOccurred())
		})

		It("migrates data from fromStore to toStore", func() {
			Expect(toStore.CreateInstanceDetailsCallCount()).To(Equal(2))
			id1, serviceInstance1 := toStore.CreateInstanceDetailsArgsForCall(0)
			id2, serviceInstance2 := toStore.CreateInstanceDetailsArgsForCall(1)
			Expect([]string{id1, id2}).To(ConsistOf([]string{"123", "456"}))
			Expect([]string{serviceInstance1.ServiceID, serviceInstance2.ServiceID}).To(ConsistOf([]string{"some-service-1", "some-service-2"}))
		})
	})

	Context("when there are binding details in fromStore", func() {
		BeforeEach(func() {
			fromStore.RetrieveAllBindingDetailsReturns(map[string]brokerapi.BindDetails{
				"123": brokerapi.BindDetails{AppGUID: "some-app-1"},
				"456": brokerapi.BindDetails{AppGUID: "some-app-2"},
			}, nil)
			err := migrationObj.Migrate(fromStore, toStore)
			Expect(err).NotTo(HaveOccurred())
		})

		It("migrates data from fromStore to toStore", func() {
			Expect(toStore.CreateBindingDetailsCallCount()).To(Equal(2))
			id1, bindDetails1 := toStore.CreateBindingDetailsArgsForCall(0)
			id2, bindDetails2 := toStore.CreateBindingDetailsArgsForCall(1)
			Expect([]string{id1, id2}).To(ConsistOf([]string{"123", "456"}))
			Expect([]string{bindDetails1.AppGUID, bindDetails2.AppGUID}).To(ConsistOf([]string{"some-app-1", "some-app-2"}))
		})
	})

	Context("when the migration is complete", func() {
		var err error

		JustBeforeEach(func() {
			err = migrationObj.Migrate(fromStore, toStore)
		})

		It("calls retire on the SQL store", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(fromStore.RetireCallCount()).To(Equal(1))
		})

		Context("when the retire call fails", func() {
			BeforeEach(func() {
				fromStore.RetireReturns(errors.New("retire-failed"))
			})

			It("returns the error from the store", func() {
				Expect(err).To(MatchError(errors.New("retire-failed")))
			})
		})
	})
})