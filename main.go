package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/InteractiveLecture/id-extractor"
	"github.com/InteractiveLecture/middlewares/jwtware"
	"github.com/InteractiveLecture/pgmapper"
	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	config := pgmapper.DefaultConfig()
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
	log.Println("listening on 8000")
	http.ListenAndServe(":8000", r)
}

func addObjectHandler(mapper *pgmapper.Mapper) http.Handler {
	result := func(w http.ResponseWriter, r *http.Request) {
		entity := make(map[string]interface{})
		err := json.NewDecoder(r.Body).Decode(&entity)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		err = mapper.Execute("insert into object_identities(id,parent_object,owner) values(%v)", entity["id"], entity["parent"])
		if err != nil {
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
			//TODO Database function missing
			err = mapper.Execute("SELECT insert_bulk_permissions(%v)", objectId, entity["create_permission"], entity["read_permission"], entity["update_permission"], entity["delete_permission"], ids)
		} else {
			err = mapper.Execute("insert into acl_entries(object_id,sid,create_permission,read_permission,update_permission,delete_permission) values($1,$2,$3,$4,$5,$6) ON CONFLICT (object_id,sid) UPDATE SET create_permission = $3, read_permission = $4, update_permission = $5, delete_permission = $6 where sid = $2 AND object_id = $1", objectId, entity["sid"], entity["create_permission"], entity["read_permission"], entity["update_permission"], entity["delete_permission"])
		}
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	return http.Handler(http.HandlerFunc(result))
}
