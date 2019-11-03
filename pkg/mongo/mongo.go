package mongo

import (
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"github.com/globalsign/mgo"
)

var session *mgo.Session
var _dbName string

func Init(host, user, pwd, dbname string) error {
	if user != "" {
		return InitCluster(host, user, pwd, dbname)
	} else {
		return InitDev(host, dbname)
	}
}

func InitCluster(host, user, pwd, dbname string) error {
	_dbName = dbname
	hostParts := strings.Split(host, "-")
	hostPre := hostParts[0]
	hostSuff := hostParts[1]
	hosts := []string{
		fmt.Sprintf("%s-shard-00-00-%s:27017", hostPre, hostSuff),
		fmt.Sprintf("%s-shard-00-01-%s:27017", hostPre, hostSuff),
		fmt.Sprintf("%s-shard-00-02-%s:27017", hostPre, hostSuff),
	}

	var err error
	dialInfo := &mgo.DialInfo{
		Addrs:    hosts,
		Username: user,
		Password: pwd,
	}
	tlsConfig := &tls.Config{}
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
		return conn, err
	}
	if session, err = mgo.DialWithInfo(dialInfo); err != nil {
		return err
	}
	return nil
}

func InitDev(host, dbname string) error {
	_dbName = dbname
	var err error
	if session, err = mgo.Dial(host); err != nil {
		return err
	}
	return nil
}

func NewClient(host, user, pwd string) (*mgo.Session, error) {
	if user == "" {
		return mgo.Dial(host)
	} else {
		hostParts := strings.Split(host, "-")
		hostPre := hostParts[0]
		hostSuff := hostParts[1]
		hosts := []string{
			fmt.Sprintf("%s-shard-00-00-%s:27017", hostPre, hostSuff),
			fmt.Sprintf("%s-shard-00-01-%s:27017", hostPre, hostSuff),
			fmt.Sprintf("%s-shard-00-02-%s:27017", hostPre, hostSuff),
		}

		dialInfo := &mgo.DialInfo{
			Addrs:    hosts,
			Username: user,
			Password: pwd,
		}
		tlsConfig := &tls.Config{}
		dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
			conn, err := tls.Dial("tcp", addr.String(), tlsConfig)
			return conn, err
		}
		return mgo.DialWithInfo(dialInfo)
	}
}

func C(collectionName string) (*mgo.Session, *mgo.Collection) {
	s := session.Copy()
	return s, s.DB(_dbName).C(collectionName)
}

func Upsert(c *mgo.Collection, result, query, update interface{}) error {
	var err error
	if result == nil {
		_, err = c.Upsert(query, update)
	} else {
		change := mgo.Change{
			Update:    update,
			Upsert:    true,
			ReturnNew: true,
		}
		_, err = c.Find(query).Apply(change, result)
	}
	return err
}

func UpdateAll(c *mgo.Collection, query, update interface{}) error {
	_, err := c.UpdateAll(query, update)
	return err
}

func Insert(c *mgo.Collection, document interface{}) error {
	err := c.Insert(document)
	return err
}

// First optional arg is Fields
// Second optional arg is slice of sort strings, ie. []string{"price", "-created_at"}
func Find(c *mgo.Collection, result, query interface{}, args ...interface{}) error {
	q := c.Find(query)
	if args != nil {
		if len(args) > 0 && args[0] != nil {
			q = q.Select(args[0])
		}
		if len(args) > 1 && args[1] != nil {
			q = q.Sort(args[1].([]string)...)
		}
	}
	if err := q.All(result); err != nil {
		return err
	}
	return nil
}

func FindOne(c *mgo.Collection, result, query interface{}) error {
	if err := c.Find(query).One(result); err != nil {
		return err
	}
	return nil
}

func Aggregate(c *mgo.Collection, result, pipe interface{}) error {
	if err := c.Pipe(pipe).All(result); err != nil {
		return err
	}
	return nil
}

func AggregateOne(c *mgo.Collection, result, pipe interface{}) error {
	if err := c.Pipe(pipe).One(result); err != nil {
		return err
	}
	return nil
}

func Remove(c *mgo.Collection, query interface{}) error {
	if _, err := c.RemoveAll(query); err != nil {
		return err
	}
	return nil
}

func CreateIndexKey(c *mgo.Collection, key ...string) error {
	return c.EnsureIndexKey(key...)
}
