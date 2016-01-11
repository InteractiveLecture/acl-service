package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"io"
	"log"
	"net/http"

	"github.com/InteractiveLecture/id-extractor"
	"github.com/InteractiveLecture/middlewares/jwtware"
	"github.com/InteractiveLecture/pgmapper"
	"github.com/gorilla/mux"
)

func main() {
	dbHost := flag.String("dbhost", "localhost", "the database host")
	dbPort := flag.Int("dbport", 5432, "the database port")
	dbUser := flag.String("dbuser", "aclapp", "the database user")
	dbSsl := flag.Bool("dbssl", false, "database ssl config")
	dbName := flag.String("dbname", "acl", "the database name")
	dbPassword := flag.String("dbpass", "", "database password")
	flag.Parse()
	config := pgmapper.DefaultConfig()
	config.Host = *dbHost
	config.Port = *dbPort
	config.User = *dbUser
	config.Ssl = *dbSsl
	config.Database = *dbName
	config.Password = *dbPassword
	r := mux.NewRouter()
	mapper, err := pgmapper.New(config)
	if err != nil {
		log.Fatal(err)
	}
	objectIdExtractor := idextractor.MuxIdExtractor("objectId")
	userIdExtractor := idextractor.MuxIdExtractor("userId")
	r.Methods("POST").Path("/objects").Handler(jwtware.New(addObjectHandler(mapper)))
	r.Methods("DELETE").Path("/objects/{objectId}").Handler(jwtware.New(deleteObjectHandler(mapper, objectIdExtractor)))
	r.Methods("GET").Path("/objects/{objectId}/permissions/{userId}").Handler(jwtware.New(getPermissionsHandler(mapper, objectIdExtractor, userIdExtractor)))
	r.Methods("PUT").Path("/objects/{objectId}/permissions").Handler(jwtware.New(upsertPermissionsHandler(mapper, objectIdExtractor)))
	r.Methods("PUT").Path("/sids/{sid}/permissions").Handler(jwtware.New(upsertMultiplePermissionsHandler(mapper, idextractor.MuxIdExtractor("sid"))))
	log.Println("listening on 8080")
	http.ListenAndServe(":8080", r)
}

func addObjectHandler(mapper *pgmapper.Mapper) http.Handler {
	result := func(w http.ResponseWriter, r *http.Request) {
		entity := make(map[string]interface{})
		err := json.NewDecoder(r.Body).Decode(&entity)
		if err != nil {
			log.Println("error while decoding json: ", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = mapper.Execute("insert into object_identities(id,parent_object) values(%v)", entity["id"], entity["parent"])
		if err != nil {
			log.Println("error while insertiing object into database: ", err)
			w.WriteHeader(http.StatusBadRequest)
		}
	}
	return http.Handler(http.HandlerFunc(result))
}

func deleteMultipleObjectsHandler(mapper *pgmapper.Mapper) http.Handler {
	result := func(w http.ResponseWriter, r *http.Request) {
		ids, ok := r.URL.Query()["oid"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
		}
		err := mapper.Execute("SELECT delete_objects(%v)", ids)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
	return http.Handler(http.HandlerFunc(result))
}

func deleteObjectHandler(mapper *pgmapper.Mapper, extractor idextractor.Extractor) http.Handler {
	result := func(w http.ResponseWriter, r *http.Request) {
		objectId, err := extractor(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, err = mapper.ExecuteRaw("delete from object_identities where id = $1", objectId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	return http.Handler(http.HandlerFunc(result))
}

func getPermissionsHandler(mapper *pgmapper.Mapper, objectIdExtractor idextractor.Extractor, userIdExtractor idextractor.Extractor) http.Handler {
	result := func(w http.ResponseWriter, r *http.Request) {
		objectId, err := objectIdExtractor(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		userId, err := userIdExtractor(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		result, err := mapper.QueryIntoBytes("SELECT get_permissions(%v)", objectId, userId)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, err = io.Copy(w, bytes.NewReader(result))
		if err != nil {
			log.Println(err)
		}
	}
	return http.Handler(http.HandlerFunc(result))
}

func upsertMultiplePermissionsHandler(mapper *pgmapper.Mapper, sidIdExtractor idextractor.Extractor) http.Handler {
	result := func(w http.ResponseWriter, r *http.Request) {
		sid, err := sidIdExtractor(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ids, ok := r.URL.Query()["oid"]
		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		permissions := make(map[string]interface{})
		err = json.NewDecoder(r.Body).Decode(&permissions)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = mapper.Execute("SELECT insert_bulk_sid_permissions(%v)", sid, permissions["create_permission"], permissions["read_permission"], permissions["update_permission"], permissions["delete_permission"], ids)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}
	return http.Handler(http.HandlerFunc(result))
}

func upsertPermissionsHandler(mapper *pgmapper.Mapper, objectIdExtractor idextractor.Extractor) http.Handler {
	result := func(w http.ResponseWriter, r *http.Request) {
		objectId, err := objectIdExtractor(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ids, ok := r.URL.Query()["sid"]
		entity := make(map[string]interface{})
		err = json.NewDecoder(r.Body).Decode(&entity)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if ok {
			err = mapper.Execute("SELECT insert_bulk_permissions(%v)", objectId, entity["create_permission"], entity["read_permission"], entity["update_permission"], entity["delete_permission"], ids)
		} else {
			_, err = mapper.ExecuteRaw("insert into acl_entries(object_id,sid,create_permission,read_permission,update_permission,delete_permission) values($1,$2,$3,$4,$5,$6) ON CONFLICT (object_id,sid) DO UPDATE SET create_permission = $3, read_permission = $4, update_permission = $5, delete_permission = $6 where acl_entries.sid = $2 AND acl_entries.object_id = $1", objectId, entity["sid"], entity["create_permission"], entity["read_permission"], entity["update_permission"], entity["delete_permission"])
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	return http.Handler(http.HandlerFunc(result))
}
