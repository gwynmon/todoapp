package mongo

import (
	"context"
	"errors"
	"todoapp/internal/entity"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type NoteRepo struct {
	coll *mongo.Collection
}

func NewNoteRepo(db *mongo.Database) *NoteRepo {
	return &NoteRepo{coll: db.Collection("notes")}
}

func (r *NoteRepo) Create(ctx context.Context, note *entity.Note) error {
	res, err := r.coll.InsertOne(ctx, note)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		note.ID = oid
	}
	return nil
}

func (r *NoteRepo) GetByTaskID(ctx context.Context, taskID int) ([]entity.Note, error) {
	filter := bson.D{{Key: "task_id", Value: taskID}}
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})

	cursor, err := r.coll.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var notes []entity.Note
	if err := cursor.All(ctx, &notes); err != nil {
		return nil, err
	}
	return notes, nil
}

func (r *NoteRepo) GetByID(ctx context.Context, noteID bson.ObjectID) (*entity.Note, error) {
	var note entity.Note
	err := r.coll.FindOne(ctx, bson.D{{Key: "_id", Value: noteID}}).Decode(&note)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, entity.ErrNoteNotFound
		}
		return nil, err
	}
	return &note, nil
}

func (r *NoteRepo) Delete(ctx context.Context, noteID bson.ObjectID) error {
	res, err := r.coll.DeleteOne(ctx, bson.D{{Key: "_id", Value: noteID}})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return entity.ErrNoteNotFound
	}
	return nil
}
