#!/bin/sh

cd util/test
go test -v -cover
cd ../..

cd repository/test
go test -v -cover
cd ../..

cd router/test
go test -v -cover
cd ../..