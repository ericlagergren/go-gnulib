void* null = ((void*)0);

struct fts
{
	struct ftsent **fts_array;
};

struct ftsent
{
	const char* name;
	struct ftsent *fts_link;
};

struct ftsent* sort(struct fts *sp, struct ftsent *head, int nitems);

int
main(void) {
	struct fts fb;

	struct ftsent fsb[5];
	fsb[4] = (struct ftsent) { "four",  null    };
	fsb[3] = (struct ftsent) { "three", &fsb[4] };
	fsb[2] = (struct ftsent) { "two",   &fsb[3] };
	fsb[1] = (struct ftsent) { "one",   &fsb[2] };
	fsb[0] = (struct ftsent) { "zero",  &fsb[1] };

	sort(&fb, &fsb[0], 5);

	return 0;
}

struct ftsent* sort(struct fts *sp, struct ftsent *head, int nitems) {
	struct ftsent **ap, *p;

    for (ap = sp->fts_array, p = head; p; p = p->fts_link) {
		*ap++ = p;
	}

    for (head = *(ap = sp->fts_array); --nitems; ++ap) {
		ap[0]->fts_link = ap[1];
	}

    ap[0]->fts_link = null;
    return head;
}