import React, {useEffect, useState} from "react";
import { useOutletContext } from "react-router-dom";
import {UserHeaderWithContext} from "./userHeaderWithContext";
import {auth} from "../../../../lib/api";
import {CredentialsShowModal, CredentialsTable} from "../../../../lib/components/auth/credentials";
import useUser from "../../../../lib/hooks/user";
import {ConfirmationButtonWithContext} from "../../../../lib/components/modals";
import {
    ActionGroup,
    ActionsBar,
    AlertError,
    RefreshButton
} from "../../../../lib/components/controls";
import {useRouter} from "../../../../lib/hooks/router";

const EMAIL_TRUNCATION_THRESHOLD_LENGTH = 40;


const UserCredentialsList = ({ userId, after, onPaginate }) => {
    const {user} = useUser();
    const [refresh, setRefresh] = useState(false);
    const [createError, setCreateError] = useState(null);
    const [createdKey, setCreatedKey] = useState(null);

    const createKey = () => {
        return auth.createCredentials(userId)
            .catch(err => {
                setCreateError(err);
            }).then(key => {
                setCreateError(null);
                setRefresh(!refresh);
                return key;
            });
    };
    const content = (
            <>
                {createError && <AlertError error={createError}/>}
                <CredentialsTable
                    userId={userId}
                    currentAccessKey={(user) ? user.accessKeyId : ""}
                    refresh={refresh}
                    after={after}
                    onPaginate={onPaginate}
                />
            </>
        );

    const getMsg = (email) => (
        <span>
                Create new credentials for user{" "}
            <br/>
            <strong
                className={`d-inline-block
                            text-nowrap
                            overflow-hidden
                            text-truncate
                            align-bottom
                            ${email.length > EMAIL_TRUNCATION_THRESHOLD_LENGTH ? "w-75" : ""}`}
                title={email}
            >
                {email}
            </strong>
            ?
        </span>
    );
    return (
        <>
            <UserHeaderWithContext userId={userId} page={'credentials'}/>

            <ActionsBar>
                <ActionGroup orientation="left">
                    <ConfirmationButtonWithContext
                        userId={userId}
                        variant="success"
                        modalVariant="success"
                        msg={getMsg}
                        onConfirm={hide => {
                            createKey()
                                .then(key => { setCreatedKey(key) })
                                .finally(hide)
                        }}>
                        Create Access Key
                    </ConfirmationButtonWithContext>
                </ActionGroup>

                <ActionGroup orientation="right">
                    <RefreshButton onClick={() => setRefresh(!refresh)}/>
                </ActionGroup>
            </ActionsBar>

            <div className="mt-2">

                <CredentialsShowModal
                    credentials={createdKey}
                    show={(!!createdKey)}
                    onHide={() => { setCreatedKey(null) }}/>

                {content}
            </div>
        </>
    );
}

const UserCredentialsContainer = () => {
    const router = useRouter();
    const { after } = router.query;
    const { userId } = router.params;
    return (!userId) ? <></> : <UserCredentialsList
        userId={userId}
        after={(after) ? after : ""}
        onPaginate={after => router.push({pathname: '/auth/users/:userId/credentials', query: {after}, params: {userId}})}
    />;
};

const UserCredentialsPage = () => {
    const {setActiveTab} = useOutletContext();
    useEffect(() => setActiveTab("users"), [setActiveTab]);
    return <UserCredentialsContainer/>;
};

export default UserCredentialsPage;