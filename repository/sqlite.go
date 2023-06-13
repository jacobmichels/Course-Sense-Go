package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	coursesense "github.com/jacobmichels/Course-Sense-Go"
	"github.com/jacobmichels/Course-Sense-Go/config"
	_ "modernc.org/sqlite"
)

var _ coursesense.Repository = SQLiteRepository{}

type SQLiteRepository struct {
	db  *sql.DB
	cfg config.SQLite
}

// creates a new repository backed by sqlite
// returns an error if the connection cannot be established or if a ping fails
func newSQLiteRepository(ctx context.Context, cfg config.SQLite) (SQLiteRepository, error) {
	// open connection
	db, err := sql.Open("sqlite", cfg.ConnectionString)
	if err != nil {
		return SQLiteRepository{}, fmt.Errorf("failed to open connection to sqlite: %w", err)
	}

	// check connection
	err = db.PingContext(ctx)
	if err != nil {
		return SQLiteRepository{}, fmt.Errorf("failed to ping db: %w", err)
	}

	// perform migrations
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return SQLiteRepository{}, fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://migrations/sqlite", "sqlite", driver)
	if err != nil {
		return SQLiteRepository{}, fmt.Errorf("failed to create migration: %w", err)
	}

	err = m.Up()
	if err != nil && err.Error() != "no change" {
		return SQLiteRepository{}, fmt.Errorf("failed to execute migrations: %w", err)
	}

	return SQLiteRepository{db, cfg}, nil
}

func (r SQLiteRepository) AddWatcher(ctx context.Context, section coursesense.Section, watcher coursesense.Watcher) error {
	txCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	tx, err := r.db.BeginTx(txCtx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// first insert the course
	course_id, err := persistCourse(txCtx, tx, section.Course)
	if err != nil {
		return fmt.Errorf("failed to persist course: %w", err)
	}

	// then insert the section
	section_id, err := persistSection(txCtx, tx, section, course_id)
	if err != nil {
		return fmt.Errorf("failed to persist section: %w", err)
	}

	// finally we can insert the watcher and commit the transaction
	err = persistWatcher(txCtx, tx, watcher, section_id)
	if err != nil {
		return fmt.Errorf("failed to persist watcher: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// insert a course into sqlite if needed
// returns the course_id
func persistCourse(txCtx context.Context, tx *sql.Tx, course coursesense.Course) (int, error) {
	// check if identical course already exists in db
	var course_id int
	err := tx.QueryRowContext(txCtx, "SELECT id FROM courses WHERE code=$1 AND department=$2", course.Code, course.Department).Scan(&course_id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		// if it doesn't exist, insert it and return the new rowid
		res, err := tx.ExecContext(txCtx, "INSERT INTO courses (code, department) VALUES ($1, $2)", course.Code, course.Department)
		if err != nil {
			return 0, fmt.Errorf("insert statement failed: %w", err)
		}

		course_id, err := res.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to fetch inserted course id: %w", err)
		}
		return int(course_id), nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to check if course already exists in database: %w", err)
	}

	// simply return the rowid if the course does exist in the db
	return course_id, nil
}

// insert a section into sqlite if needed
// returns the section_id
func persistSection(txCtx context.Context, tx *sql.Tx, section coursesense.Section, course_id int) (int, error) {
	// check if identical section already exists in db
	var section_id int
	err := tx.QueryRowContext(txCtx, "SELECT id FROM sections WHERE code=$1 AND term=$2 AND course_id=$3", section.Code, section.Term, course_id).Scan(&section_id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		// if it doesn't exist, insert it and return the new rowid
		res, err := tx.ExecContext(txCtx, "INSERT INTO sections (code, term, course_id) VALUES ($1, $2, $3)", section.Code, section.Term, course_id)
		if err != nil {
			return 0, fmt.Errorf("insert statement failed: %w", err)
		}

		section_id, err := res.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("failed to fetch inserted section id: %w", err)
		}
		return int(section_id), nil
	} else if err != nil {
		return 0, fmt.Errorf("failed to check if section already exists in database: %w", err)
	}

	// simply return the rowid if the course does exist in the db
	return section_id, nil
}

// insert a watcher into sqlite if needed
func persistWatcher(txCtx context.Context, tx *sql.Tx, watcher coursesense.Watcher, section_id int) error {
	// check if identical watcher already exists in db
	var watcher_id int
	err := tx.QueryRowContext(txCtx, "SELECT id FROM watchers WHERE email=$1 AND section_id=$2", watcher.Email, section_id).Scan(&watcher_id)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		// if it doesn't exist, insert it
		_, err := tx.ExecContext(txCtx, "INSERT INTO watchers (email, section_id) VALUES ($1, $2)", watcher.Email, section_id)
		if err != nil {
			return fmt.Errorf("insert statement failed: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check if watcher already exists in database: %w", err)
	}
	return nil
}

func (r SQLiteRepository) GetWatchedSections(ctx context.Context) ([]coursesense.Section, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT courses.code, courses.department, sections.code, sections.term FROM sections left join courses on sections.course_id=courses.id")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sections from the db: %w", err)
	}

	var sections []coursesense.Section

	defer rows.Close()
	for rows.Next() {
		var section coursesense.Section

		if err := rows.Scan(&section.Course.Code, &section.Course.Department, &section.Code, &section.Term); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		sections = append(sections, section)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return sections, nil
}

func (r SQLiteRepository) GetWatchers(ctx context.Context, section coursesense.Section) ([]coursesense.Watcher, error) {
	// get section_id for section in question
	var section_id int
	err := r.db.QueryRowContext(ctx, "SELECT sections.id FROM sections left join courses on sections.course_id=courses.id WHERE sections.code=$1 AND sections.term=$2 AND courses.department=$3 AND courses.code=$4", section.Code, section.Term, section.Course.Department, section.Course.Code).Scan(&section_id)
	if err != nil {
		return nil, fmt.Errorf("failed to get section_id from db: %w", err)
	}

	// then get the emails
	rows, err := r.db.QueryContext(ctx, "SELECT email FROM watchers WHERE section_id=$1", section_id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch relevant watchers from db: %w", err)
	}

	var watchers []coursesense.Watcher

	defer rows.Close()
	for rows.Next() {
		var watcher coursesense.Watcher

		if err := rows.Scan(&watcher.Email); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		watchers = append(watchers, watcher)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate rows: %w", err)
	}

	return watchers, nil
}

func (r SQLiteRepository) Cleanup(ctx context.Context, section coursesense.Section) error {
	// start by removing watchers
	// get section_id for section in question
	var section_id int
	err := r.db.QueryRowContext(ctx, "SELECT sections.id FROM sections left join courses on sections.course_id=courses.id WHERE sections.code=$1 AND sections.term=$2 AND courses.department=$3 AND courses.code=$4", section.Code, section.Term, section.Course.Department, section.Course.Code).Scan(&section_id)
	if err != nil {
		return fmt.Errorf("failed to get section_id from db: %w", err)
	}

	_, err = r.db.ExecContext(ctx, "DELETE FROM watchers WHERE section_id=$1", section_id)
	if err != nil {
		return fmt.Errorf("failed to execute delete command: %w", err)
	}

	// get the id of the course referenced by the section
	var course_id int
	err = r.db.QueryRowContext(ctx, "SELECT courses.id FROM courses LEFT JOIN sections ON courses.id=sections.course_id WHERE courses.code=$1 AND courses.department=$2 AND sections.code=$3 AND sections.term=$4", section.Course.Code, section.Course.Department, section.Code, section.Term).Scan(&course_id)
	if err != nil {
		return fmt.Errorf("failed to fetch course_id from db: %w", err)
	}

	// remove the related row in the section table
	_, err = r.db.ExecContext(ctx, "DELETE FROM sections WHERE id=$1", section_id)
	if err != nil {
		return fmt.Errorf("failed to delete section: %w", err)
	}

	// check if there are any other sections of the same course in the table
	var count int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM sections WHERE course_id=$1", course_id).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to count same course sections: %w", err)
	}

	// if no other sections reference this course, we can safely delete it
	if count == 0 {
		log.Info().Msgf("deleting course %v %v", section.Course.Department, section.Course.Code)
		_, err = r.db.ExecContext(ctx, "DELETE FROM courses WHERE id=$1", course_id)
		if err != nil {
			return fmt.Errorf("failed to delete course: %w", err)
		}
	}

	return nil
}
