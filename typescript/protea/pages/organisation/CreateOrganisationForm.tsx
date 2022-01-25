import { useMutation } from '@apollo/client'
import { Button, Routes, TextField } from 'components'
import {
  CreateOrganisationMutation,
  CreateOrganisationMutationVariables,
  CreateOrganisationDocument
} from 'generated/types'
import { useRouter } from 'next/router'
import React, { FC } from 'react'
import { useForm } from 'react-hook-form'

type CreateOrganisationInputs = {
  name: string
}

export const CreateOrganisationForm: FC = () => {
  const [saveOrg] = useMutation<
    CreateOrganisationMutation,
    CreateOrganisationMutationVariables
  >(CreateOrganisationDocument)
  const router = useRouter()
  const {
    register,
    handleSubmit,
    setError,
    formState: { errors, isValid }
  } = useForm<CreateOrganisationInputs>()

  const onSubmit = (inputs: CreateOrganisationInputs) => {
    saveOrg({
      variables: inputs
    })
      .then((val) => {
        return router.push({
          pathname: Routes.organisationOverview,
          query: { orgId: val.data?.createOrganisation.organisation?.id }
        })
      })
      .catch((e) => {
        // TODO handle appropriately when we have error catching service
        setError('name', {
          type: 'manual',
          message: 'Something went wrong.'
        })
      })
  }

  return (
    <form
      className='flex min-w-full flex-col space-y-4'
      onSubmit={handleSubmit(onSubmit)}
    >
      <TextField
        {...register('name', { required: 'Organisation name is required.' })}
        id='name'
        label='Organisation name'
        type='text'
        isValid={isValid}
        errorMessage={errors.name?.message}
      />

      <div className='flex items-center justify-end'>
        <Button type='submit'>Create organisation</Button>
      </div>
    </form>
  )
}
