""""agent_removed_type_and_default"

Revision ID: a9d0e1c399ad
Revises: d965e4c90e30
Create Date: 2017-05-29 09:24:34.900381

"""
from alembic import op
import sqlalchemy as sa
from sqlalchemy.sql import table, column
from sqlalchemy.orm import Session
from sqlalchemy import update
# revision identifiers, used by Alembic.
revision = 'a9d0e1c399ad'
down_revision = 'd965e4c90e30'
branch_labels = None
depends_on = None


def upgrade():
    # ### commands auto generated by Alembic - please adjust! ###
    op.drop_column('agent', 'type')
    op.drop_column('agent', 'default')
    # ### end Alembic commands ###


def downgrade():
    # ### commands auto generated by Alembic - please adjust! ###
    op.add_column(
        'agent', sa.Column(
            'default', sa.Boolean(), nullable=True, default=True))
    op.add_column(
        'agent', sa.Column(
            'type', sa.Enum('solo', 'multi', name='agent_types'),
            nullable=False, default='solo'))

    bind = op.get_bind()
    session = Session(bind=bind)
    agent_table = table('agent', column('type'), column('default'))

    update_type = update(agent_table).values(type='solo')
    update_default = update(agent_table).values(default=True)
    try:
        session.execute(update_type)
        session.execute(update_default)
        session.commit()
    finally:
        session.close()
    # ### end Alembic commands ###
