/*=============================================================================
#     FileName: defaultdispatcher.go
#         Desc: default dispatcher
#       Author: sunminghong
#        Email: allen.fantasy@gmail.com
#     HomePage: http://weibo.com/5d13
#      Version: 0.0.1
#   LastChange: 2013-06-06 14:37:57
#      History:
=============================================================================*/
package gate

import (
    "strconv"
    "strings"
    "math/rand"
    . "github.com/sunminghong/letsgo/helper"
    . "github.com/sunminghong/letsgo/log"
)

type LGDefaultDispatcher struct {
    messageCodemaps map[int]LGSliceInt
    removeGrids map[int]int
}

func LGNewDispatcher() *LGDefaultDispatcher {
    r := &LGDefaultDispatcher{make(map[int]LGSliceInt),make(map[int]int)}
    return r
}

func (r *LGDefaultDispatcher)Init() {
    r.messageCodemaps = make(map[int]LGSliceInt)
}

func (r *LGDefaultDispatcher) Add(gridID int, gcodes *string) {
    cs := strings.Replace(*gcodes," ","",-1)
    if len(cs) ==0 {
        r.addDisp(gridID,0)
        return
    }

    codes:= strings.Split(cs,",")
    LGTrace("add disp",codes)
    for _,p_ := range codes {
        p := strings.Trim(p_," ")
        if len(p) == 0 {
            continue
        }
        gcode, err := strconv.Atoi(p)
        if err ==nil {
            r.addDisp(gridID,gcode)
        }
    }
    LGTrace("messagecodemaps1:",r.messageCodemaps)
}

func (r *LGDefaultDispatcher) Remove(gridID int) {
    r.removeGrids[gridID] = 1
    //LGTrace("removegrids:",r.removeGrids)
}

func (r *LGDefaultDispatcher) addDisp(gridID int, gcode int) {
    dises,ok := r.messageCodemaps[gcode]
    if ok {
        r.messageCodemaps[gcode] = append(dises,gridID)
        return
    }

    r.messageCodemaps[gcode] = LGSliceInt{gridID}
}

func (r *LGDefaultDispatcher) Dispatch(messageCode int) (gridID int,ok bool) {
    gcode := r.groupCode(messageCode)

    gridIDArr,ok := r.messageCodemaps[gcode]
    LGTrace("gridIDArr=%v,gcode:%v,maps:%v",gridIDArr,gcode,r.messageCodemaps)
    if !ok {
        gcode = 0
        gridIDArr,ok = r.messageCodemaps[gcode]
        //LGTrace("gridIDArr2,gcode:",gridIDArr,gcode)
    }

    if !ok {
        return 0,false
    }

    l := len(gridIDArr)
    if l == 0 {
        return 0,false
    }

    i := rand.Intn(l)
    //LGTrace("rand:",l,i)
    for i<l {
        gridID = gridIDArr[i]
        _,ok := r.removeGrids[gridID]
        //LGTrace("removeGrids: ",r.removeGrids)
        if ok {
            //LGTrace("this grid is already down",gridID)

            gridIDArr.RemoveAtIndex(i)
            r.messageCodemaps[gcode] = gridIDArr

            //LGTrace("removed ",i,":",gridIDArr)
            l = len(gridIDArr)
            if l == 0 {
                break
            }

            if i >= l {
                i = 0
            }
            continue
        }

        //LGTrace(
        //    "dispatcher Handler func messageCode,messageCode,gridID:",
        //messageCode,gcode,gridID,gridIDArr)

        return gridID,true
    }

    return 0,false
}

//将协议编号分组以供Dispatch决策用那个Grid 来处理
func (r *LGDefaultDispatcher) groupCode(messageCode int) int {
    return int(messageCode / 100)
}

