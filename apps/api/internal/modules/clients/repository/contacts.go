package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"riders.agency/pulse/api/internal/platform/authz"
)

func (c *Clients) ListContacts(ctx context.Context, clientID, uid uuid.UUID) ([]ContactRow, error) {
	if _, err := c.clientCompany(ctx, clientID, uid, authz.PermRead); err != nil {
		return nil, err
	}
	rows, err := c.P.Query(ctx, `
SELECT client_contact_id::text, client_id::text, COALESCE(first_name,''), COALESCE(last_name,''),
       COALESCE(middle_name,''), COALESCE(position,''), COALESCE(email,''), COALESCE(phone,''),
       COALESCE(telegram,''), preferred_channel, COALESCE(comment,''), is_primary, status
  FROM client_contacts
 WHERE client_id = $1 AND status <> 'deleted'
 ORDER BY is_primary DESC, last_name, first_name`, clientID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ContactRow{}
	for rows.Next() {
		var item ContactRow
		if err := rows.Scan(&item.ClientContactID, &item.ClientID, &item.FirstName, &item.LastName, &item.MiddleName,
			&item.Position, &item.Email, &item.Phone, &item.Telegram, &item.PreferredChannel, &item.Comment,
			&item.IsPrimary, &item.Status); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (c *Clients) CreateContactTx(ctx context.Context, clientID, uid uuid.UUID, in ContactInput) (string, error) {
	companyID, err := c.clientCompany(ctx, clientID, uid, authz.PermWrite)
	if err != nil {
		return "", err
	}
	in.PreferredChannel = normalizeDefault(in.PreferredChannel, "email")
	in.Status = normalizeDefault(in.Status, "active")
	id := uuid.New()
	tx, err := c.transaction()
	if err != nil {
		return "", err
	}
	if in.IsPrimary {
		_, _ = tx.Exec(ctx, `UPDATE client_contacts SET is_primary = false WHERE client_id = $1`, clientID.String())
	}
	_, err = tx.Exec(ctx, `
INSERT INTO client_contacts (
  client_contact_id, client_id, company_id, first_name, last_name, middle_name, position,
  email, phone, telegram, preferred_channel, comment, is_primary, status)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
		id.String(), clientID.String(), companyID.String(), nullStr(in.FirstName), nullStr(in.LastName),
		nullStr(in.MiddleName), nullStr(in.Position), nullStr(in.Email), nullStr(in.Phone), nullStr(in.Telegram),
		in.PreferredChannel, nullStr(in.Comment), in.IsPrimary, in.Status)
	if err != nil {
		return "", err
	}
	return id.String(), nil
}

func (c *Clients) PatchContactTx(ctx context.Context, clientID, contactID, uid uuid.UUID, patch ContactPatch) error {
	if _, err := c.clientCompany(ctx, clientID, uid, authz.PermWrite); err != nil {
		return err
	}
	tx, err := c.transaction()
	if err != nil {
		return err
	}
	if patch.IsPrimary != nil && *patch.IsPrimary {
		_, _ = tx.Exec(ctx, `UPDATE client_contacts SET is_primary = false WHERE client_id = $1`, clientID.String())
	}
	tag, err := tx.Exec(ctx, `
UPDATE client_contacts SET
  first_name = COALESCE($3, first_name),
  last_name = COALESCE($4, last_name),
  middle_name = COALESCE($5, middle_name),
  position = COALESCE($6, position),
  email = COALESCE($7, email),
  phone = COALESCE($8, phone),
  telegram = COALESCE($9, telegram),
  preferred_channel = COALESCE($10, preferred_channel),
  comment = COALESCE($11, comment),
  is_primary = COALESCE($12, is_primary),
  status = COALESCE($13, status),
  updated_at = now()
WHERE client_id = $1 AND client_contact_id = $2`,
		clientID.String(), contactID.String(), optTrim(patch.FirstName), optTrim(patch.LastName), optTrim(patch.MiddleName),
		optTrim(patch.Position), optTrim(patch.Email), optTrim(patch.Phone), optTrim(patch.Telegram),
		optTrim(patch.PreferredChannel), optTrim(patch.Comment), patch.IsPrimary, optStr(patch.Status))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (c *Clients) DeleteContact(ctx context.Context, clientID, contactID, uid uuid.UUID) error {
	if _, err := c.clientCompany(ctx, clientID, uid, authz.PermWrite); err != nil {
		return err
	}
	_, err := c.P.Exec(ctx, `
UPDATE client_contacts SET status = 'deleted', updated_at = now()
 WHERE client_id = $1 AND client_contact_id = $2`, clientID.String(), contactID.String())
	return err
}
