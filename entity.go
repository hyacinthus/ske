package main

import (
	"encoding/json"
	"net/http"

	"github.com/hyacinthus/ske/ske"
	"github.com/hyacinthus/x/xerr"
	"github.com/labstack/echo"
	nsq "github.com/nsqio/go-nsq"
)

// ============= 业务部分 ==============

func findEntityByID(id string) (*ske.Entity, error) {
	var r = new(ske.Entity)
	if err := db.Where("id = ?", id).First(r).Error; err != nil {
		return nil, err
	}
	return r, nil
}

// ============= 事件处理部分 ==============

// ReceiveEntity 处理收到的对象
// Topic: entity_new Channel: ske
// Body: Entity
func ReceiveEntity(msg *nsq.Message) error {
	entity := new(ske.Entity)
	err := json.Unmarshal(msg.Body, entity)
	if err != nil {
		log.WithError(err).Errorf("接收到的消息格式错误: %+v", msg)
		return err
	}
	log.Info(entity)
	return nil
}

// ============= REST 部分 ==============

// createEntity 新建实体
func createEntity(c echo.Context) error {
	// 输入
	var r = new(ske.Entity)
	if err := c.Bind(r); err != nil {
		return err
	}
	// 校验
	if r.Title == "" {
		return xerr.New(400, "BadRequest", "Empty title")
	}
	// 保存
	if err := db.Create(r).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, r)
}

// updateEntity 更新实体
func updateEntity(c echo.Context) error {
	// 获取URL中的ID
	id := c.Param("id")
	var n = new(ske.EntityUpdate)
	if err := c.Bind(n); err != nil {
		return err
	}
	old, err := findEntityByID(id)
	if err != nil {
		return err
	}
	// 利用指针检查是否有请求这个字段
	if n.Title != nil {
		if *n.Title == "" {
			return xerr.New(400, "BadRequest", "Empty title")
		}
		old.Title = *n.Title
	}

	if err := db.Save(old).Error; err != nil {
		return err
	}

	return c.JSON(http.StatusOK, old)
}

// deleteEntity 删除实体
func deleteEntity(c echo.Context) error {
	id := c.Param("id")
	// 删除数据库对象
	if err := db.Where("id = ?", id).Delete(&ske.Entity{}).Error; err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// getEntity 获取实体
func getEntity(c echo.Context) error {
	id := c.Param("id")
	r, err := findEntityByID(id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, r)
}

// getEntitys 获取实体列表
func getEntitys(c echo.Context) error {
	// 提前make可以让查询没有结果的时候返回空列表
	var ns = make([]*ske.Entity, 0)
	// 分页信息
	limit := c.Get("limit").(int)
	offset := c.Get("offset").(int)
	err := db.Order("updated_at desc").
		Offset(offset).Limit(limit).Find(&ns).Error
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, ns)
}
