// Copyright (C) MongoDB, Inc. 2014-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

package db

import (
	"context"
	"testing"

	"github.com/mongodb/mongo-tools-common/options"
	"github.com/mongodb/mongo-tools-common/testtype"
	. "github.com/smartystreets/goconvey/convey"
	"go.mongodb.org/mongo-driver/bson"
)

func TestBufferedBulkInserterInserts(t *testing.T) {
	testtype.SkipUnlessTestType(t, testtype.IntegrationTestType)

	var bufBulk *BufferedBulkInserter

	auth := DBGetAuthOptions()
	ssl := DBGetSSLOptions()
	Convey("With a valid session", t, func() {
		opts := options.ToolOptions{
			Connection: &options.Connection{
				Port: DefaultTestPort,
			},
			SSL:  &ssl,
			Auth: &auth,
		}
		provider, err := NewSessionProvider(opts)
		So(provider, ShouldNotBeNil)
		So(err, ShouldBeNil)
		session, err := provider.GetSession()
		So(session, ShouldNotBeNil)
		So(err, ShouldBeNil)

		Convey("using a test collection and a doc limit of 3", func() {
			testCol := session.Database("tools-test").Collection("bulk1")
			bufBulk = NewBufferedBulkInserter(testCol, 3, false)
			So(bufBulk, ShouldNotBeNil)

			Convey("inserting 10 documents into the BufferedBulkInserter", func() {
				flushCount := 0
				for i := 0; i < 10; i++ {
					So(bufBulk.Insert(bson.D{}), ShouldBeNil)
					if bufBulk.docCount%3 == 0 {
						flushCount++
					}
				}

				Convey("should have flushed 3 times with one doc still buffered", func() {
					So(flushCount, ShouldEqual, 3)
					So(bufBulk.byteCount, ShouldBeGreaterThan, 0)
					So(bufBulk.docCount, ShouldEqual, 1)
				})
			})
		})

		Convey("using a test collection and a doc limit of 1", func() {
			testCol := session.Database("tools-test").Collection("bulk2")
			bufBulk = NewBufferedBulkInserter(testCol, 1, false)
			So(bufBulk, ShouldNotBeNil)

			Convey("inserting 10 documents into the BufferedBulkInserter and flushing", func() {
				for i := 0; i < 10; i++ {
					So(bufBulk.Insert(bson.D{}), ShouldBeNil)
				}
				So(bufBulk.Flush(), ShouldBeNil)

				Convey("should have no docs buffered", func() {
					So(bufBulk.docCount, ShouldEqual, 0)
					So(bufBulk.byteCount, ShouldEqual, 0)
				})
			})
		})

		Convey("using a test collection and a doc limit of 1000", func() {
			testCol := session.Database("tools-test").Collection("bulk3")
			bufBulk = NewBufferedBulkInserter(testCol, 100, false)
			So(bufBulk, ShouldNotBeNil)

			Convey("inserting 1,000,000 documents into the BufferedBulkInserter and flushing", func() {

				for i := 0; i < 1000000; i++ {
					bufBulk.Insert(bson.M{"_id": i})
				}
				So(bufBulk.Flush(), ShouldBeNil)

				Convey("should have inserted all of the documents", func() {
					count, err := testCol.CountDocuments(context.Background(), bson.M{})
					So(err, ShouldBeNil)
					So(count, ShouldEqual, 1000000)

					// test values
					testDoc := bson.M{}
					result := testCol.FindOne(nil, bson.M{"_id": 477232})
					err = result.Decode(&testDoc)
					So(err, ShouldBeNil)
					So(testDoc["_id"], ShouldEqual, 477232)
					result = testCol.FindOne(nil, bson.M{"_id": 999999})
					err = result.Decode(&testDoc)
					So(err, ShouldBeNil)
					So(testDoc["_id"], ShouldEqual, 999999)
					result = testCol.FindOne(nil, bson.M{"_id": 1})
					err = result.Decode(&testDoc)
					So(err, ShouldBeNil)
					So(testDoc["_id"], ShouldEqual, 1)

				})
			})
		})

		Reset(func() {
			provider.DropDatabase("tools-test")
			provider.Close()
		})
	})

}
