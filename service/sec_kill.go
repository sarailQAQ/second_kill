package service

import (
	"log"
	"sync"
	"time"
)

type User struct {
	UserId string
	GoodsId  uint
}

var OrderChan = make(chan User, 1024)

var ItemMap = make(map[uint]*Item)
var mutex = &sync.RWMutex{}


type Item struct {
	ID        uint   // 商品id
	Name      string // 名字
	Price	  int	//价格
	Total     int    // 商品总量
	Left      int    // 商品剩余数量
	IsSoldOut bool   // 是否售罄
	leftCh    chan int
	sellCh    chan int
	done      chan struct{}
	Lock      sync.Mutex
}

// TODO 写一个定时任务，每天定时从数据库加载数据到Map
func OnShelve() {
	beginTime := time.Now()
	// 获取第二天时间
	nextTime := beginTime.Add(time.Hour * 24)
	// 计算次日六点点，即商品下架的时间
	onShelveTime := time.Date(nextTime.Year(), nextTime.Month(), nextTime.Day(), 6, 0, 0, 0, nextTime.Location())
//	offShelveTime := beginTime.Add(time.Second*5)
	timer := time.NewTimer(onShelveTime.Sub(beginTime))

	<-timer.C
	initMap()
}

//从数据库加载货物信息
func initMap() {
	mutex.Lock()
	defer mutex.Unlock()
	goods,err := GoodsInit()
	if err != nil {
		log.Println("error load goods",err)
		return
	}

	for _,g := range goods {
		item := &Item{
			ID:        g.ID,
			Name:      g.Name,
			Price:     g.Price ,
			Total:     g.Num,
			Left:      g.Num,
			IsSoldOut: false,
			leftCh:    make(chan int),
			sellCh:    make(chan int),
		}
		ItemMap[item.ID] = item
	}

}


func getItem(itemId uint) *Item{
	mutex.RLock()
	defer mutex.RUnlock()
	return ItemMap[itemId]
}

//处理秒杀请求
func order() {
	for {
		user := <- OrderChan
		item := getItem(user.GoodsId)
		item.SecKilling(user.UserId)
	}
}

func (item *Item) SecKilling(userId string) {

	item.Lock.Lock()
	defer item.Lock.Unlock()
	// 等价
	// var lock = make(chan struct{}, 1}
	// lock <- struct{}{}
	// defer func() {
    // 		<- lock
    // }
	if item.IsSoldOut {
		return
	}
	item.BuyGoods(1)

	MakeOrder(userId, item.ID,1)
}

// 定时下架
func (item *Item) OffShelve() {
	beginTime := time.Now()
	// 获取第二天时间
	nextTime := beginTime.Add(time.Hour * 24)
	// 计算次日零点，即商品下架的时间
	offShelveTime := time.Date(nextTime.Year(), nextTime.Month(), nextTime.Day(), 0, 0, 0, 0, nextTime.Location())
	//offShelveTime := beginTime.Add(time.Second*5)
	timer := time.NewTimer(offShelveTime.Sub(beginTime))

	<-timer.C
	delete(ItemMap, item.ID)
	close(item.done)
	close(item.leftCh)
	close(item.sellCh)

	//定时上架
	go OnShelve()
}

// 出售商品
func (item *Item) SalesGoods() {
	for {
		select {
		case num := <-item.sellCh:
			if item.Left -= num; item.Left <= 0 {
				item.IsSoldOut = true
			}

		case item.leftCh <- item.Left:
		case <-item.Done():
			log.Println("我自闭了")
			return
		}
	}
}

func (item *Item) Done() <-chan struct{} {
	if item.done == nil {
		item.done = make(chan struct{})
	}
	d := item.done
	return d
}

func (item *Item) Monitor() {
	go item.SalesGoods()
}

// 获取剩余库存
func (item *Item) GetLeft() int {
	var left int
	left = <-item.leftCh
	return left
}

// 购买商品
func (item *Item) BuyGoods(num int) {
	item.sellCh <- num
}

//启动服务
func InitService() {
	initMap()
	for _,item := range ItemMap{
		item.Monitor()
		go item.OffShelve()
	}
	//异步处理订单
	for i := 0; i < 10; i++ {
		go order()
	}
}
