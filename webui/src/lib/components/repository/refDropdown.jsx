import React, {useEffect, useRef, useState} from "react";

import Form from "react-bootstrap/Form";
import Alert from "react-bootstrap/Alert";
import Button from "react-bootstrap/Button";
import Badge from "react-bootstrap/Badge";
import Overlay from "react-bootstrap/Overlay";
import {Col, Nav, Row} from "react-bootstrap";
import Popover from "react-bootstrap/Popover";
import ListGroup from 'react-bootstrap/ListGroup';
import {ChevronDownIcon, ChevronRightIcon, ChevronUpIcon, XIcon} from "@primer/octicons-react";

import {tags, branches, commits} from '../../api';
import {RefTypeBranch, RefTypeCommit, RefTypeTag} from "../../../constants";

const RefSelector = ({ repo, selected, selectRef, withCommits, withWorkspace, withTags, amount = 300 }) => {
    // used for ref pagination
    const [pagination, setPagination] = useState({after: "", prefix: "", amount});
    const [refList, setRefs] = useState({loading: true, payload: null, error: null});
    const [refType, setRefType] = useState(selected && selected.type || RefTypeBranch)
    useEffect(() => {
        setRefs({loading: true, payload: null, error: null});
        const fetchRefs = async () => {
            try {
                let response;
                if (refType === RefTypeTag) {
                    response = await tags.list(repo.id, pagination.prefix, pagination.after, pagination.amount);
                } else {
                    response = await branches.list(repo.id, pagination.prefix, pagination.after, pagination.amount);
                }
                setRefs({loading: false, payload: response, error: null});
            } catch (error) {
                setRefs({loading: false, payload: null, error: error});
            }
        };
        fetchRefs();
    }, [refType, repo.id, pagination])

    // used for commit listing
    const initialCommitList = {branch: selected, commits: null, loading: false};
    const [commitList, setCommitList] = useState(initialCommitList);


    const form = (
        <div className="ref-filter-form">
            <Form onSubmit={e => { e.preventDefault(); }}>
                <Form.Control type="text" placeholder={refType === RefTypeTag ? "Filter tags" : "Filter branches"} onChange={(e)=> {
                    setPagination({
                        amount,
                        after: "",
                        prefix: e.target.value
                    })
                }}/>
            </Form>
        </div>
    );
    const refTypeNav = withTags && <Nav variant="tabs" onSelect={setRefType} activeKey={refType} className="mt-2">
        <Nav.Item>
            <Nav.Link eventKey={"branch"}>Branches</Nav.Link>
        </Nav.Item>
        <Nav.Item>
            <Nav.Link eventKey={"tag"}>Tags</Nav.Link>
        </Nav.Item>
    </Nav>

    if (refList.loading) {
        return  (
            <div className="ref-selector">
                {form}
                {refTypeNav}
                <p>Loading...</p>
            </div>
        );
    }

    if (refList.error) {
        return  (
            <div className="ref-selector">
                {form}
                {refTypeNav}
                <Alert variant="danger">{refList.error}</Alert>
            </div>
        );
    }

    if (commitList.commits !== null) {
        return (
            <CommitList
                withWorkspace={withWorkspace}
                commits={commitList.commits}
                branch={commitList.branch}
                selectRef={selectRef}
                reset={() => {
                    setCommitList(initialCommitList);
                }}/>
        );
    }


    const results = refList.payload.results;

    return (
        <div className="ref-selector">
            {form}
            {refTypeNav}
            <div className="ref-scroller">
                {(results && results.length > 0) ? (
                    <>
                        <ListGroup as="ul" className="ref-list">
                            {results.map(namedRef => (
                                <RefEntry key={namedRef.id} repo={repo} refType={refType} namedRef={namedRef.id} selectRef={selectRef} selected={selected} withCommits={refType !== RefTypeTag && withCommits} logCommits={async () => {
                                    const data = await commits.log(repo.id, namedRef.id)
                                    setCommitList({...commitList, branch: namedRef.id, commits: data.results});
                                }}/>
                            ))}
                        </ListGroup>
                        <Paginator results={refList.payload.results} pagination={refList.payload.pagination} from={pagination.after} onPaginate={(after) => {
                            setPagination({after})
                        }}/>
                    </>
                ) : (
                    <p className="text-center mt-3"><small>No references found</small></p>
                )}

            </div>
        </div>
    );
};

const CommitList = ({ commits, selectRef, reset, branch, withWorkspace }) => {
    const getMessage = commit => {
        if (!commit.message) {
            return 'repository epoch';
        }

        if (commit.message.length > 60) {
            return commit.message.substr(0, 40) + '...';
        }

        return commit.message;
    };

    return (
        <div className="ref-selector">
            <h5>{branch}</h5>
            <div className="ref-scroller">
                <ul className="list-group ref-list">
                    {(withWorkspace) ? (
                        <li className="list-group-item" key={branch}>
                            <Button variant="link" onClick={() => {
                                selectRef({id: branch, type: RefTypeBranch});
                            }}><em>{branch}{'\''}s Workspace (uncommitted changes)</em></Button>
                        </li>
                    ) : (<span/>)}
                    {commits.map(commit => (
                        <li className="list-group-item" key={commit.id}>
                            <Button variant="link" onClick={() => {
                                selectRef({id: commit.id, type: RefTypeCommit});
                            }}>{getMessage(commit)} </Button>
                            <div className="actions">
                                <Badge variant="light">{commit.id.substr(0, 12)}</Badge>
                            </div>
                        </li>
                    ))}
                </ul>
                <p className="ref-paginator">
                    <Button variant="link" size="sm" onClick={reset}>Back</Button>
                </p>
            </div>
        </div>
    );
};

const RefEntry = ({repo, namedRef, refType, selectRef, selected, logCommits, withCommits}) => {
    return (
        <ListGroup.Item as="li" key={namedRef}>
            <Row className="align-items-center">
                <Col
                    title={namedRef}
                    className="text-nowrap overflow-hidden text-truncate"
                >
                    {!!selected && namedRef === selected.id ?
                        <strong>{namedRef}</strong> :
                        <Button
                            variant="link"
                            onClick={() => selectRef({ id: namedRef, type: refType })}
                        >
                            {namedRef}
                        </Button>
                    }
                </Col>
                <Col xs="auto" className="actions d-flex align-items-center">
                    {refType === RefTypeBranch && namedRef === repo.default_branch && (
                        <Badge variant="info" className="mr-2">
                            Default
                        </Badge>
                    )}
                    {withCommits && (
                        <Button onClick={logCommits} size="sm" variant="link">
                            <ChevronRightIcon />
                        </Button>
                    )}
                </Col>
            </Row>
        </ListGroup.Item>
    );
};

const Paginator = ({ pagination, onPaginate, results, from }) => {
    const next = (results.length) ? results[results.length-1].id : "";

    if (!pagination.has_more && from === "") return (<span/>);

    return (
        <p className="ref-paginator">
            {(from !== "") ?
                (<Button  size={"sm"} variant="link" onClick={() => { onPaginate(""); }}>Reset</Button>) :
                (<span/>)
            }
            {' '}
            {(pagination.has_more) ?
                (<Button size={"sm"} variant="link" onClick={() => { onPaginate(next); }}>Next...</Button>) :
                (<span/>)
            }
        </p>
    );
};

const RefDropdown = ({ repo, selected, selectRef, onCancel, variant="light", prefix = '', emptyText = '', withCommits = true, withWorkspace = true, withTags = true }) => {

    const [show, setShow] = useState(false);
    const target = useRef(null);


    const popover = (
        <Overlay target={target.current} show={show} placement="bottom" rootClose={true} onHide={() => setShow(false)}>
            <Popover className="ref-popover">
                <Popover.Body>
                    <RefSelector
                        repo={repo}
                        withCommits={withCommits}
                        withWorkspace={withWorkspace}
                        withTags={withTags}
                        selected={selected}
                        selectRef={(ref) => {
                            selectRef(ref);
                            setShow(false);
                        }}/>
                </Popover.Body>
            </Popover>
        </Overlay>
    );

    const cancelButton = (!!onCancel && !!selected) ? (<Button onClick={() => {
        setShow(false);
        onCancel();
    }} variant={variant}><XIcon/></Button>) : (<span/>);

    if (!selected) {
        return (
            <>
                <Button ref={target} variant={variant} onClick={()=> { setShow(!show) }}>
                    {emptyText} {show ? <ChevronUpIcon/> : <ChevronDownIcon/>}
                </Button>
                {cancelButton}
                {popover}
            </>
        );
    }

    const showId = (ref) => {
        if (!ref)
            return ''
        if (ref.type === RefTypeCommit)
            return ref.id.substr(0, 12)
        return ref.id
    }

    const title = prefix + (!!selected) ? `${prefix} ${selected.type}: ` : '';
    return (
        <>
            <Button ref={target}
                    variant={variant}
                    onClick={() => setShow(!show)}
                    style={{ maxWidth: '250px' }}
                    title={showId(selected)}
                    className="d-inline-flex align-items-center"
            >
                <span className="text-truncate">
                    {title}
                    <strong>{showId(selected)}</strong>
                </span>
                <span className="ms-1">
                    {show ? <ChevronUpIcon /> : <ChevronDownIcon />}
                </span>
            </Button>
            {cancelButton}
            {popover}
        </>
    );
};

export default RefDropdown;
