""""cloud_agent_assignment"

Revision ID: 116f5bf4c7dc
Revises: 0f7376d5f910
Create Date: 2018-10-05 15:11:38.104615

"""
from alembic import op
import sqlalchemy as sa
from sqlalchemy.dialects import mysql
from sqlalchemy.orm import Session
from sqlalchemy.sql import table, column
from sqlalchemy import update, String, select, Integer
import datetime

# revision identifiers, used by Alembic.
revision = '116f5bf4c7dc'
down_revision = '0f7376d5f910'
branch_labels = None
depends_on = None


def upgrade():
    # ### commands auto generated by Alembic - please adjust! ###
    op.alter_column('cloud_agent', 'api_url',
                    existing_type=String(256),
                    nullable=True)
    op.alter_column('cloud_agent', 'iscsi_address',
                    existing_type=String(256),
                    nullable=True)
    op.drop_constraint('cloud_agent_ibfk_1', 'cloud_agent', type_='foreignkey')
    op.drop_index('uc_customer_id_api_url_deleted_at',
                  table_name='cloud_agent')
    op.create_foreign_key('cloud_agent_ibfk_1', 'cloud_agent', 'customer',
                          ['customer_id'], ['id'])
    op.add_column('cloud_volume', sa.Column('assigned_cloud_agent_id',
                                            sa.String(length=36),
                                            nullable=True))
    op.create_foreign_key('volume_ibfk_cloud_agent', 'cloud_volume',
                          'cloud_agent', ['assigned_cloud_agent_id'], ['id'])
    # ### end Alembic commands ###


def downgrade():
    # ### commands auto generated by Alembic - please adjust! ###
    bind = op.get_bind()
    session = Session(bind=bind)
    try:
        device_table = table('cloud_agent', column('deleted_at', Integer()),
                             column('id', String(36)))
        stmt = select([device_table]).where(device_table.c.deleted_at == 0)
        devices = session.execute(stmt)
        now = datetime.datetime.utcnow().timestamp()
        for i, device in enumerate(devices):
            stmt = update(device_table).values(deleted_at=now+i).where(
                device['id'] == device_table.c.id)
            session.execute(stmt)
        session.commit()
    finally:
        session.close()
    op.drop_constraint('volume_ibfk_cloud_agent', 'cloud_volume',
                       type_='foreignkey')
    op.drop_constraint('cloud_agent_ibfk_1', 'cloud_agent', type_='foreignkey')
    op.drop_column('cloud_volume', 'assigned_cloud_agent_id')
    op.create_unique_constraint(
        'uc_customer_id_api_url_deleted_at', 'cloud_agent',
        ['customer_id', 'api_url', 'deleted_at'])
    op.create_foreign_key('cloud_agent_ibfk_1', 'cloud_agent', 'customer',
                          ['customer_id'], ['id'])
    op.alter_column('cloud_agent', 'iscsi_address',
                    existing_type=String(256),
                    nullable=False)
    op.alter_column('cloud_agent', 'api_url',
                    existing_type=String(256),
                    nullable=False)
    # ### end Alembic commands ###
