/*=============================================================================
#     FileName: profilingtool.go
#       Author: sunminghong, allen.fantasy@gmail.com, http://weibo.com/5d13
#         Team: http://1201.us
#   LastChange: 2013-12-13 14:50:22
#      History:
=============================================================================*/


/*
记录cpu 性能
*/
package helper

/*
import (
        "fmt"
        "io"
        "log"
        "os"
        "runtime"
        "runtime/debug"
        "runtime/pprof"
        "strconv"
        "sync/atomic"
        "syscall"
        "time"
)

var heapProfileCounter int32
var startTime = time.Now()
var pid int

func init() {
        pid = os.Getpid()
}

func LGStartCPUProfile() {
        f, err := os.Create("cpu-" + strconv.Itoa(pid) + ".pprof")
        if err != nil {
                log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
}

func LGStopCPUProfile() {
        pprof.StopCPUProfile()
}

func LGStartBlockProfile(rate int) {
        runtime.SetBlockProfileRate(rate)
}

func LGStopBlockProfile() {
        filename := "block-" + strconv.Itoa(pid) + ".pprof"
        f, err := os.Create(filename)
        if err != nil {
                log.Fatal(err)
        }
        if err = pprof.Lookup("block").WriteTo(f, 0); err != nil {
                log.Fatalf(" can't write %s: %s", filename, err)
        }
        f.Close()
}

func LGSetMemProfileRate(rate int) {
        runtime.MemProfileRate = rate
}

func LGGC() {
        runtime.GC()
}

func LGDumpHeap() {
        filename := "heap-" + strconv.Itoa(pid) + "-" + strconv.Itoa(int(atomic.AddInt32(&heapProfileCounter, 1))) + ".pprof"
        f, err := os.Create(filename)
        if err != nil {
                fmt.Fprintf(os.Stderr, "testing: %s", err)
                return
        }
        if err = pprof.WriteHeapProfile(f); err != nil {
                fmt.Fprintf(os.Stderr, "testing: can't write %s: %s", filename, err)
        }
        f.Close()
}

func showSystemStat(interval time.Duration, count int) {

        usage1 := &syscall.Rusage{}
        var lastUtime int64
        var lastStime int64

        counter := 0
        for {

                //http://man7.org/linux/man-pages/man3/vtimes.3.html
                syscall.Getrusage(syscall.RUSAGE_SELF, usage1)

                utime := int64(usage1.Utime.Sec*1000000000) + int64(usage1.Utime.Usec)
                stime := int64(usage1.Stime.Sec*1000000000) + int64(usage1.Stime.Usec)
                userCPUUtil := float64(utime-lastUtime) * 100 / float64(interval)
                sysCPUUtil := float64(stime-lastStime) * 100 / float64(interval)
                memUtil := usage1.Maxrss * 1024

                lastUtime = int64(utime)
                lastStime = int64(stime)

                if counter > 0 {
                        fmt.Printf("cpu: %3.2f%% us  %3.2f%% sy, mem:%s \n", userCPUUtil, sysCPUUtil, toH(uint64(memUtil)))
                }

                counter += 1
                if count >= 1 && count < counter {
                        return
                }
                time.Sleep(interval)
        }

}

func LGShowSystemStat(seconds int) {
        go func() {
                interval := time.Duration(seconds) * time.Second
                showSystemStat(interval, 0)
        }()
}

func LGPrintSystemStats() {
        interval := time.Duration(1) * time.Second
        showSystemStat(interval, 1)
}

func LGShowGCStat() {
        go func() {
                var numGC int64

                interval := time.Duration(100) * time.Millisecond
                gcstats := &debug.GCStats{PauseQuantiles: make([]time.Duration, 100)}
                memStats := &runtime.MemStats{}
                for {
                        debug.ReadGCStats(gcstats)
                        if gcstats.NumGC > numGC {
                                runtime.ReadMemStats(memStats)

                                printGC(memStats, gcstats, os.Stdout)
                                numGC = gcstats.NumGC
                        }
                        time.Sleep(interval)
                }
        }()
}

func LGPrintGCSummary() {
        LGFprintGCSummary(os.Stdout)
}

func LGFprintGCSummary(output io.Writer) {
        memStats := &runtime.MemStats{}
        runtime.ReadMemStats(memStats)
        gcstats := &debug.GCStats{PauseQuantiles: make([]time.Duration, 100)}
        debug.ReadGCStats(gcstats)

        printGC(memStats, gcstats, output)
}

func printGC(memStats *runtime.MemStats, gcstats *debug.GCStats, output io.Writer) {

        if gcstats.NumGC > 0 {
                lastPause := gcstats.Pause[0]
                elapsed := time.Now().Sub(startTime)
                overhead := float64(gcstats.PauseTotal) / float64(elapsed) * 100
                allocatedRate := float64(memStats.TotalAlloc) / elapsed.Seconds()

                fmt.Fprintf(output, "NumGC:%d Pause:%s Pause(Avg):%s Overhead:%3.2f%% Alloc:%s Sys:%s Alloc(Rate):%s/s Histogram:%s %s %s \n",
                        gcstats.NumGC,
                        toS(lastPause),
                        toS(avg(gcstats.Pause)),
                        overhead,
                        toH(memStats.Alloc),
                        toH(memStats.Sys),
                        toH(uint64(allocatedRate)),
                        toS(gcstats.PauseQuantiles[94]),
                        toS(gcstats.PauseQuantiles[98]),
                        toS(gcstats.PauseQuantiles[99]))
        } else {
                // while GC has disabled
                elapsed := time.Now().Sub(startTime)
                allocatedRate := float64(memStats.TotalAlloc) / elapsed.Seconds()

                fmt.Fprintf(output, "Alloc:%s Sys:%s Alloc(Rate):%s/s\n",
                        toH(memStats.Alloc),
                        toH(memStats.Sys),
                        toH(uint64(allocatedRate)))
        }
}

func avg(items []time.Duration) time.Duration {
        var sum time.Duration
        for _, item := range items {
                sum += item
        }
        return time.Duration(int64(sum) / int64(len(items)))
}

// human readable format
func toH(bytes uint64) string {
        switch {
        case bytes < 1024:
                return fmt.Sprintf("%dB", bytes)
        case bytes < 1024*1024:
                return fmt.Sprintf("%.2fK", float64(bytes)/1024)
        case bytes < 1024*1024*1024:
                return fmt.Sprintf("%.2fM", float64(bytes)/1024/1024)
        default:
                return fmt.Sprintf("%.2fG", float64(bytes)/1024/1024/1024)
        }
}

// short string format
func toS(d time.Duration) string {

        u := uint64(d)
        if u < uint64(time.Second) {
                switch {
                case u == 0:
                        return "0"
                case u < uint64(time.Microsecond):
                        return fmt.Sprintf("%.2fns", float64(u))
                case u < uint64(time.Millisecond):
                        return fmt.Sprintf("%.2fus", float64(u)/1000)
                default:
                        return fmt.Sprintf("%.2fms", float64(u)/1000/1000)
                }
        } else {
                switch {
                case u < uint64(time.Minute):
                        return fmt.Sprintf("%.2fs", float64(u)/1000/1000/1000)
                case u < uint64(time.Hour):
                        return fmt.Sprintf("%.2fm", float64(u)/1000/1000/1000/60)
                default:
                        return fmt.Sprintf("%.2fh", float64(u)/1000/1000/1000/60/60)
                }
        }
}
*/
