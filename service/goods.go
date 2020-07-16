package service

import (
	"errors"
	"github.com/jinzhu/gorm"
	"log"
	"summerCourse/model"
	"sync"
)

type Goods struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Num   int    `json:"num"`
}

// 添加商品
func AddGoods(good Goods) error {
	// TODO
	g := model.Goods{
		Model: gorm.Model{
			ID:        good.ID,
		},
		Name:  good.Name,
		Price: good.Price,
		Num:   good.Num,
	}
	err := g.AddGoods()
	if err != nil {
		log.Println(err)
		return err
	}
	//直接从map中读取，节省开销
	if getItem(good.ID).ID > 0 {
		return errors.New("Good has exited!")
	}
	mutex.Lock()
	defer mutex.Unlock()
	ItemMap[g.ID] = &Item{
		ID:        good.ID,
		Name:      good.Name,
		Total:     good.Num,
		Left:      good.Num,
		IsSoldOut: false,
		leftCh:    make(chan int),
		sellCh:    make(chan int),
		done:      nil,
		Lock:      sync.Mutex{},
	}
	return nil
}

//直接从itemMap读取商品信息
func SelectGoods() (goods []Goods) {
	mutex.RLock()
	for _,item := range ItemMap{
		var good  = Goods{
			ID:    item.ID,
			Name:  item.Name,
			Price: item.Price,
			Num:   item.Left,
		}
		goods = append(goods, good)
	}
	mutex.RUnlock()
	return
}

//从mysql读取商品信息
func GoodsInit() (goods []Goods,err error) {
		_goods, err := model.SelectGoods()
		if err != nil {
			log.Printf("Error get goods info. Error: %s", err)
			return
		}
		for _, v := range _goods {
			good := Goods{
				ID:    v.ID,
				Name:  v.Name,
				Price: v.Price,
				Num:   v.Num,
			}
			goods = append(goods, good)
		}
		return
}