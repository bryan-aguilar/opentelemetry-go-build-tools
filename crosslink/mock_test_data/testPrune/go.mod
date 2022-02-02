module go.opentelemetry.io/build-tools/crosslink/testroot

go 1.17

require (
    go.opentelemetry.io/build-tools/crosslink/testroot/testA v1.0.0
    go.opentelemetry.io/build-tools/crosslink/testroot/testB v1.0.0
    go.opentelemetry.io/build-tools/crosslink/testroot/testC v1.0.0
    go.opentelemetry.io/build-tools/crosslink/testroot/testD v1.0.0
    go.opentelemetry.io/build-tools/crosslink/testroot/testE v1.0.0
    go.opentelemetry.io/build-tools/crosslink/testroot/testF v1.0.0
)


replace go.opentelemetry.io/build-tools/crosslink/testroot/testA => ./testA

replace go.opentelemetry.io/build-tools/crosslink/testroot/testB => ./testB

replace go.opentelemetry.io/build-tools/crosslink/testroot/testC => ./testC

replace go.opentelemetry.io/build-tools/crosslink/testroot/testD => ./testD

replace go.opentelemetry.io/build-tools/crosslink/testroot/testE => ./testE

replace go.opentelemetry.io/build-tools/crosslink/testroot/testF => ./testF

// mock transitive dependencies
// should be in the required replace field in modInfo struct
replace go.opentelemetry.io/build-tools/crosslink/testroot/testG => ./testG

replace go.opentelemetry.io/build-tools/crosslink/testroot/testH => ./testH

// test repositories that aren't under the root module do not get removed
// even if they are not under the dependency graph. We do not make this type of destructive change.
// This can possibly be changed if the namign convention rules are changed.
replace go.opentelemetry.io/not-a-real-module/testFoo => ./testFoo

replace go.opentelemetry.io/fake-module/ => ./fake-module

// test that parent modules do not get removed
replace go.opentelemetry.io/build-tools/multimod => ../multimod

// inter repository replace statements should remain
replace foo.opentelemetery.io/bar => ../bar

// Inter-repository with a transitive dependency. 
// Not an issue with pruning, this would be an issue with inserting replace statements. Pruning would see
// that this is in the list of required replace statements and would not remove. 
replace go.opentelemetry.io/build-tools/crosslink/testroot/testK => ../crosslinkcopy/testK

// should be pruned
replace go.opentelemetry.io/build-tools/crosslink/testroot/testI => ./testI

replace go.opentelemetry.io/build-tools/crosslink/testroot/testJ => ./testJ








