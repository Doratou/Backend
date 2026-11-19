-- +goose Up
CREATE TYPE checklist_item_type AS ENUM ('GROUP', 'PERSONAL');

CREATE TABLE checklist_items (
    id BIGSERIAL PRIMARY KEY,
    group_id BIGINT NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
    title VARCHAR(200) NOT NULL CHECK (length(btrim(title)) > 0),
    type checklist_item_type NOT NULL,
    created_by_member_id BIGINT NOT NULL,
    assigned_to_member_id BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ,
    CONSTRAINT fk_checklist_items_creator_group FOREIGN KEY (created_by_member_id, group_id)
        REFERENCES group_members(id, group_id) ON DELETE RESTRICT,
    CONSTRAINT fk_checklist_items_assignee_group FOREIGN KEY (assigned_to_member_id, group_id)
        REFERENCES group_members(id, group_id) ON DELETE RESTRICT,
    CONSTRAINT chk_checklist_items_personal_unassigned
        CHECK (type = 'GROUP' OR assigned_to_member_id IS NULL)
);

CREATE INDEX idx_checklist_items_group ON checklist_items (group_id, id DESC)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_checklist_items_assignee ON checklist_items (assigned_to_member_id, group_id)
    WHERE deleted_at IS NULL AND assigned_to_member_id IS NOT NULL;

-- +goose StatementBegin
CREATE FUNCTION release_checklist_items_on_member_departure() RETURNS TRIGGER AS $$
BEGIN
    UPDATE checklist_items
    SET assigned_to_member_id = NULL, updated_at = clock_timestamp()
    WHERE group_id = NEW.group_id AND assigned_to_member_id = NEW.id AND deleted_at IS NULL;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_release_checklist_items_on_member_departure
AFTER UPDATE OF status, left_at ON group_members
FOR EACH ROW WHEN (
    OLD.status = 'ACTIVE' AND OLD.left_at IS NULL
    AND (NEW.status <> 'ACTIVE' OR NEW.left_at IS NOT NULL)
)
EXECUTE FUNCTION release_checklist_items_on_member_departure();

-- +goose Down
DROP TRIGGER trg_release_checklist_items_on_member_departure ON group_members;
DROP FUNCTION release_checklist_items_on_member_departure();
DROP TABLE checklist_items;
DROP TYPE checklist_item_type;
